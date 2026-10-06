package tests

import (
	"testing"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database"
)

// TestHarnessServiceContextMapping mengunci pemetaan service → bounded context
// di harness. Tabelnya sengaja ditulis ulang di sini (bukan diimpor) supaya
// pergeseran di databases.go benar-benar terdeteksi: kalau salah, service
// di-wire ke database yang bukan miliknya dan test tetap hijau karena database
// yang salah itu kosong.
func TestHarnessServiceContextMapping(t *testing.T) {
	want := map[string]string{
		"auth":        ContextIdentity,
		"user":        ContextIdentity,
		"role":        ContextIdentity,
		"category":    ContextCatalog,
		"product":     ContextCatalog,
		"merchant":    ContextMerchant,
		"cashier":     ContextMerchant,
		"order":       ContextSales,
		"order_item":  ContextSales,
		"transaction": ContextSales,
	}

	for svc, wantCtx := range want {
		got, ok := serviceContext[svc]
		if !ok {
			t.Errorf("serviceContext tidak memuat service %q", svc)
			continue
		}
		if got != wantCtx {
			t.Errorf("serviceContext[%q] = %q, ingin %q", svc, got, wantCtx)
		}
	}
	for svc := range serviceContext {
		if _, ok := want[svc]; !ok {
			t.Errorf("serviceContext memuat service tak terduga %q — tambahkan ke tabel ini", svc)
		}
	}
}

// TestHarnessCollectionOwnership memastikan setiap koleksi dimiliki tepat satu
// context, dan dua arah konsisten: collectionContext adalah kebalikan dari
// contextCollections.
func TestHarnessCollectionOwnership(t *testing.T) {
	want := map[string]string{
		"users":              ContextIdentity,
		"roles":              ContextIdentity,
		"user_roles":         ContextIdentity,
		"refresh_tokens":     ContextIdentity,
		"reset_tokens":       ContextIdentity,
		"categories":         ContextCatalog,
		"products":           ContextCatalog,
		"merchants":          ContextMerchant,
		"merchant_documents": ContextMerchant,
		"cashiers":           ContextMerchant,
		"orders":             ContextSales,
		"order_items":        ContextSales,
		"transactions":       ContextSales,
	}

	for coll, wantCtx := range want {
		got, ok := collectionContext[coll]
		if !ok {
			t.Errorf("collectionContext tidak memuat koleksi %q", coll)
			continue
		}
		if got != wantCtx {
			t.Errorf("collectionContext[%q] = %q, ingin %q", coll, got, wantCtx)
		}
	}
	for coll := range collectionContext {
		if _, ok := want[coll]; !ok {
			t.Errorf("collectionContext memuat koleksi tak terduga %q — tambahkan ke tabel ini", coll)
		}
	}

	// contextCollections harus menjadi sumber kebenaran: tiap koleksi yang
	// dideklarasikan di sana harus menunjuk balik ke context pemiliknya.
	for ctxName, colls := range contextCollections {
		for _, coll := range colls {
			if got := collectionContext[coll]; got != ctxName {
				t.Errorf("contextCollections[%q] memuat %q, tapi collectionContext[%q] = %q",
					ctxName, coll, coll, got)
			}
		}
	}
}

// TestHarnessContextDatabases mengunci nama database fisik per context. Harness
// harus memakai nama yang sama dengan produksi (pkg/database/names.go) —
// kalau tidak, test menguji database yang tidak pernah dipakai service.
func TestHarnessContextDatabases(t *testing.T) {
	want := map[string]string{
		ContextIdentity: database.IdentityDB,
		ContextCatalog:  database.CatalogDB,
		ContextMerchant: database.MerchantDB,
		ContextSales:    database.SalesDB,
		ContextEmail:    database.EmailDB,
	}

	for ctxName, wantDB := range want {
		got, ok := contextDatabase[ctxName]
		if !ok {
			t.Errorf("contextDatabase tidak memuat context %q", ctxName)
			continue
		}
		if got != wantDB {
			t.Errorf("contextDatabase[%q] = %q, ingin %q", ctxName, got, wantDB)
		}
	}
	for ctxName := range contextDatabase {
		if _, ok := want[ctxName]; !ok {
			t.Errorf("contextDatabase memuat context tak terduga %q — tambahkan ke tabel ini", ctxName)
		}
	}
}
