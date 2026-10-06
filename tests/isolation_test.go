package tests

import (
	"testing"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// TestContextsAreIsolated membuktikan isolasi runtime yang sebelumnya tidak
// pernah diuji: dulu harness memakai satu database testdb untuk semua service, jadi
// pemetaan DBName hanya diuji di level deklarasi (lihat db_mapping_test.go) dan
// pelanggaran lintas-context tidak pernah terlihat.
//
// Test ini menegakkan tiga hal:
//  1. setiap bounded context memakai database fisik yang berbeda;
//  2. dokumen yang ditulis ke satu context tidak terlihat dari context lain;
//  3. counter (GetNextID) juga per-context, bukan global.
func TestContextsAreIsolated(t *testing.T) {
	ts, err := SetupTestSuite(ContextSales)
	if err != nil {
		t.Fatalf("SetupTestSuite gagal: %v", err)
	}
	defer ts.Teardown()

	ctx := ts.Ctx

	// (1) Setiap context harus punya database sendiri.
	seen := make(map[string]string, len(ts.Databases))
	for ctxName, db := range ts.Databases {
		if prev, dup := seen[db.Name()]; dup {
			t.Fatalf("context %q dan %q memakai database fisik yang sama (%q)", prev, ctxName, db.Name())
		}
		seen[db.Name()] = ctxName
	}

	// (2) Dokumen milik identity tidak boleh terlihat dari sales.
	const marker = "isolation-check@example.com"
	if _, err := ts.Coll("users").InsertOne(ctx, bson.M{
		"_id":   int32(1),
		"email": marker,
	}); err != nil {
		t.Fatalf("insert users ke context identity gagal: %v", err)
	}

	visible, err := ts.DB(ContextIdentity).Collection("users").
		CountDocuments(ctx, bson.M{"email": marker})
	if err != nil {
		t.Fatalf("count di context identity gagal: %v", err)
	}
	if visible != 1 {
		t.Fatalf("users tidak terlihat di context identity (count=%d)", visible)
	}

	leaked, err := ts.DB(ContextSales).Collection("users").
		CountDocuments(ctx, bson.M{"email": marker})
	if err != nil {
		t.Fatalf("count di context sales gagal: %v", err)
	}
	if leaked != 0 {
		t.Fatalf("users bocor ke database sales (count=%d) — isolasi context tidak berjalan", leaked)
	}

	// (3) Counter harus ikut terpisah: menaikkan counter identity tidak boleh
	// menggeser counter sales untuk koleksi yang sama.
	id, err := database.GetNextID(ctx, ts.DB(ContextIdentity), "users")
	if err != nil {
		t.Fatalf("GetNextID di identity gagal: %v", err)
	}
	if id != 1 {
		t.Fatalf("counter identity.users = %d, ingin 1 (counter tidak per-context)", id)
	}

	salesID, err := database.GetNextID(ctx, ts.DB(ContextSales), "users")
	if err != nil {
		t.Fatalf("GetNextID di sales gagal: %v", err)
	}
	if salesID != 1 {
		t.Fatalf("counter sales.users = %d, ingin 1 — counter identity ikut menggeser sales", salesID)
	}
}
