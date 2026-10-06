package tests

import (
	"fmt"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// Bounded contexts. The harness models the deployment topology: every service
// owns exactly one context, and no context touches another context's
// collections directly. Production services declare their own DBName in
// service/*/cmd/main.go (see pkg/database/names.go); the harness has to know the
// same topology so it can wire each service to its own database.
//
// Sebelumnya harness memakai satu database testdb untuk semua service, sehingga
// pemetaan DBName tidak pernah diuji saat runtime dan pelanggaran lintas-context
// tidak pernah terlihat.
const (
	ContextIdentity = "identity"
	ContextCatalog  = "catalog"
	ContextMerchant = "merchant"
	ContextSales    = "sales"
	ContextEmail    = "email"
)

// contextDatabase maps a bounded context to the physical database name.
var contextDatabase = map[string]string{
	ContextIdentity: database.IdentityDB,
	ContextCatalog:  database.CatalogDB,
	ContextMerchant: database.MerchantDB,
	ContextSales:    database.SalesDB,
	ContextEmail:    database.EmailDB,
}

// contextCollections lists the domain collections each context owns. Indexes and
// counters are created per context, so a collection written to the wrong
// database is no longer silently usable.
//
// Hanya koleksi yang di-seed/dibaca test yang didaftarkan di sini. Koleksi
// infrastruktur (outbox_events per publisher, consumer_inbox, counters) tidak
// dirutekan lewat nama karena tidak unik per context — counters dibuat langsung
// di EnsureIndexes, dan outbox/inbox tidak disentuh test.
var contextCollections = map[string][]string{
	ContextIdentity: {"users", "roles", "user_roles", "refresh_tokens", "reset_tokens"},
	ContextCatalog:  {"categories", "products"},
	ContextMerchant: {"merchants", "merchant_documents", "cashiers"},
	ContextSales:    {"orders", "order_items", "transactions"},
	ContextEmail:    {},
}

// collectionContext is the inverse of contextCollections: which context owns a
// collection. Test helpers use it to write seeds to the owning database instead
// of whatever database happens to be at hand.
var collectionContext = func() map[string]string {
	m := make(map[string]string)
	for ctx, colls := range contextCollections {
		for _, c := range colls {
			m[c] = ctx
		}
	}
	return m
}()

// serviceContext maps a service to the bounded context (and therefore the
// database) it owns. Keep in sync with service/*/cmd/main.go.
var serviceContext = map[string]string{
	"auth": ContextIdentity, "user": ContextIdentity, "role": ContextIdentity,
	"category": ContextCatalog, "product": ContextCatalog,
	"merchant": ContextMerchant, "cashier": ContextMerchant,
	"order": ContextSales, "order_item": ContextSales, "transaction": ContextSales,
}

// DB returns the database that backs a bounded context.
func (ts *TestSuite) DB(context string) *mongo.Database {
	db, ok := ts.Databases[context]
	if !ok {
		panic(fmt.Sprintf("unknown bounded context %q", context))
	}
	return db
}

// DBForCollection returns the database that owns collection.
func (ts *TestSuite) DBForCollection(collection string) *mongo.Database {
	ctx, ok := collectionContext[collection]
	if !ok {
		panic(fmt.Sprintf("unknown collection %q: add it to contextCollections", collection))
	}
	return ts.DB(ctx)
}

// Coll returns a collection from the database that owns it.
func (ts *TestSuite) Coll(collection string) *mongo.Collection {
	return ts.DBForCollection(collection).Collection(collection)
}

// DBForService returns the database owned by a service.
func (ts *TestSuite) DBForService(service string) *mongo.Database {
	ctx, ok := serviceContext[service]
	if !ok {
		panic(fmt.Sprintf("unknown service %q: add it to serviceContext", service))
	}
	return ts.DB(ctx)
}

// DB returns the database that backs a bounded context.
func (s *BaseTestSuite) DB(context string) *mongo.Database {
	return s.ts.DB(context)
}

// DBForCollection returns the database that owns collection.
func (s *BaseTestSuite) DBForCollection(collection string) *mongo.Database {
	return s.ts.DBForCollection(collection)
}

// Coll returns a collection from the database that owns it.
func (s *BaseTestSuite) Coll(collection string) *mongo.Collection {
	return s.ts.Coll(collection)
}

// DBForService returns the database owned by a service.
func (s *BaseTestSuite) DBForService(service string) *mongo.Database {
	return s.ts.DBForService(service)
}
