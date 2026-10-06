package database

// Database names for each bounded context. Each context owns its collections
// exclusively; cross-context access must go through gRPC, never a shared
// MongoDB database.
//
// Nama database TIDAK diturunkan dari nama service. Setiap service menyatakan
// sendiri database yang dimilikinya lewat server.Config.DBName. Konsekuensinya:
//
//   - typo nama konstanta (mis. database.IdentytyDB) = compile error
//   - DBName kosong = gagal saat start, bukan fallback senyap
//   - paket ini tidak perlu tahu daftar nama service (tidak ada coupling
//     ke deployment topology)
//
// Lihat STUDI_KASUS_DB_CLUSTER.md §8 untuk alasan perubahan ini.
const (
	IdentityDB = "pos_identity"
	CatalogDB  = "pos_catalog"
	SalesDB    = "pos_sales"
	MerchantDB = "pos_merchant"
	EmailDB    = "pos_email"
)

// ContextPrefix returns the environment-variable prefix that points at the
// MongoDB instance dedicated to the given context database.
//
// Setiap bounded context punya server MongoDB-nya sendiri ("1 database = 1
// MongoDB"). Prefix ini adalah prefix env penuh yang dipakai
// NewMongoClientWithPrefix, jadi nilainya "MONGO_<CONTEXT>" dan key yang dibaca
// adalah MONGO_<CONTEXT>_URI / MONGO_<CONTEXT>_HOST / MONGO_<CONTEXT>_PORT.
//
// Nama yang tidak dikenal mengembalikan "MONGO" supaya pemanggil jatuh ke
// instance tunggal lama (MONGO_URI / MONGO_HOST).
func ContextPrefix(databaseName string) string {
	switch databaseName {
	case IdentityDB:
		return "MONGO_IDENTITY"
	case CatalogDB:
		return "MONGO_CATALOG"
	case SalesDB:
		return "MONGO_SALES"
	case MerchantDB:
		return "MONGO_MERCHANT"
	case EmailDB:
		return "MONGO_EMAIL"
	default:
		return "MONGO"
	}
}
