package tests

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database"
)

// expectedDBName adalah sumber kebenaran INDEPENDEN untuk pemetaan
// service → bounded context. Sengaja ditulis ulang di sini (bukan diimpor dari
// kode produksi) supaya pergeseran di kode benar-benar terdeteksi.
//
// Kenapa test ini ada:
//
//	Config.DBName diisi dengan konstanta string (database.IdentityDB, dst).
//	Semua konstanta itu SAH, jadi salah pilih — misalnya user-service memakai
//	database.SalesDB — tetap compile dan tetap jalan, hanya menulis ke database
//	yang bukan miliknya. Compiler tidak bisa menangkap kelas kesalahan ini,
//	jadi test ini yang menangkapnya.
//
// Kalau test ini gagal karena pemetaan memang sengaja diubah, ubah tabel ini
// DAN tabel kepemilikan di STUDI_KASUS_DB_CLUSTER.md §4.1.
var expectedDBName = map[string]string{
	"auth":        "IdentityDB",
	"user":        "IdentityDB",
	"role":        "IdentityDB",
	"product":     "CatalogDB",
	"category":    "CatalogDB",
	"order":       "SalesDB",
	"order_item":  "SalesDB",
	"transaction": "SalesDB",
	"merchant":    "MerchantDB",
	"cashier":     "MerchantDB",
}

// serverConfig menampung field yang diekstrak dari server.Config{...} di main.go.
type serverConfig struct {
	found       bool
	serviceName string
	dbConst     string // nama konstanta, mis. "IdentityDB" dari database.IdentityDB
}

// findServerConfig mem-parse main.go dan mengekstrak server.Config{...}.
//
// Test ini bekerja di level AST, bukan runtime, karena deklarasi DBName hanya
// ada di package main setiap service — tidak bisa diimpor test mana pun.
func findServerConfig(t *testing.T, path string) serverConfig {
	t.Helper()

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		t.Fatalf("gagal parse %s: %v", path, err)
	}

	var out serverConfig

	ast.Inspect(file, func(n ast.Node) bool {
		lit, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}

		// hanya tertarik pada server.Config{...}
		sel, ok := lit.Type.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "Config" {
			return true
		}
		if pkg, ok := sel.X.(*ast.Ident); !ok || pkg.Name != "server" {
			return true
		}

		out.found = true

		for _, elt := range lit.Elts {
			kv, ok := elt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}
			key, ok := kv.Key.(*ast.Ident)
			if !ok {
				continue
			}

			switch key.Name {
			case "DBName":
				out.dbConst = selectorTail(kv.Value)
			case "ServiceName":
				if basic, ok := kv.Value.(*ast.BasicLit); ok && basic.Kind == token.STRING {
					if s, err := strconv.Unquote(basic.Value); err == nil {
						out.serviceName = s
					}
				}
			}
		}

		return false
	})

	return out
}

// selectorTail mengembalikan nama terakhir dari sebuah selector/identifier:
// database.IdentityDB → "IdentityDB".
func selectorTail(e ast.Expr) string {
	switch v := e.(type) {
	case *ast.SelectorExpr:
		return v.Sel.Name
	case *ast.Ident:
		return v.Name
	default:
		return ""
	}
}

func serviceFromPath(path string) string {
	// ../service/<svc>/cmd/main.go → <svc>
	parts := strings.Split(filepath.ToSlash(path), "/")
	for i, p := range parts {
		if p == "service" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

// TestServiceDBMapping memastikan setiap service yang memakai server.Config
// mendeklarasikan DBName yang benar, dan ServiceName-nya cocok dengan direktorinya.
func TestServiceDBMapping(t *testing.T) {
	paths, err := filepath.Glob(filepath.Join("..", "service", "*", "cmd", "main.go"))
	if err != nil {
		t.Fatalf("glob gagal: %v", err)
	}
	if len(paths) == 0 {
		t.Fatal("tidak menemukan service/*/cmd/main.go — cwd test mungkin bukan direktori tests/")
	}

	seen := map[string]bool{}

	for _, path := range paths {
		svc := serviceFromPath(path)
		if svc == "" {
			t.Errorf("tidak bisa menyimpulkan nama service dari path %s", path)
			continue
		}

		cfg := findServerConfig(t, path)
		if !cfg.found {
			// Service tanpa server.Config (mis. email memilih DB-nya sendiri
			// lewat NewMongoDatabase langsung) — dilewati dengan sengaja.
			continue
		}

		seen[svc] = true

		t.Run(svc, func(t *testing.T) {
			want, ok := expectedDBName[svc]
			if !ok {
				t.Fatalf("%s memakai server.Config tapi belum terdaftar di expectedDBName — "+
					"tambahkan pemetaannya, atau pastikan service ini memang tidak butuh DB", path)
			}

			if cfg.dbConst == "" {
				t.Fatalf("%s: field DBName tidak ada/kosong. Tanpa DBName, server.New akan gagal saat start", path)
			}
			if cfg.dbConst != want {
				t.Errorf("%s: DBName = database.%s, ingin database.%s\n"+
					"  → salah pilih konstanta tetap compile dan tetap jalan, tapi menulis ke database yang bukan miliknya.\n"+
					"  → kalau memang sengaja, ubah expectedDBName DAN STUDI_KASUS_DB_CLUSTER.md §4.1",
					path, cfg.dbConst, want)
			}

			if wantSvc := svc + "-service"; cfg.serviceName != wantSvc {
				t.Errorf("%s: ServiceName = %q, ingin %q (indikasi main.go hasil copy-paste)",
					path, cfg.serviceName, wantSvc)
			}
		})
	}

	// Arah sebaliknya: setiap service yang diharapkan harus benar-benar ada.
	var missing []string
	for svc := range expectedDBName {
		if !seen[svc] {
			missing = append(missing, svc)
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Errorf("expectedDBName memuat service yang tidak ditemukan/di-skip: %v\n"+
			"  → kalau service dihapus/di-rename, bersihkan expectedDBName", missing)
	}
}

// TestDatabaseConstantValues mengunci nama database fisik. Tanpa ini, rename
// konstanta (mis. "pos_sales" → "pos_sales_v2") akan diam-diam menunjuk ke
// database kosong baru, tanpa error apa pun.
func TestDatabaseConstantValues(t *testing.T) {
	got := map[string]string{
		"IdentityDB": database.IdentityDB,
		"CatalogDB":  database.CatalogDB,
		"SalesDB":    database.SalesDB,
		"MerchantDB": database.MerchantDB,
		"EmailDB":    database.EmailDB,
	}
	want := map[string]string{
		"IdentityDB": "pos_identity",
		"CatalogDB":  "pos_catalog",
		"SalesDB":    "pos_sales",
		"MerchantDB": "pos_merchant",
		"EmailDB":    "pos_email",
	}

	for name, wantVal := range want {
		if got[name] != wantVal {
			t.Errorf("database.%s = %q, ingin %q", name, got[name], wantVal)
		}
	}
}
