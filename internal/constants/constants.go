package constants

// I know, we use snake case here, which is inconsistent, but I chose to use it for the constants for readiblity
const (
    ML_KEM_1024_SK_LEN = 3168
    ML_KEM_1024_PK_LEN = 1568
    ML_KEM_1024_CT_LEN = 1568

    ML_DSA_87_SK_LEN   = 4896
    ML_DSA_87_PK_LEN   = 2592
    ML_DSA_87_SIGN_LEN = 4627

    CHALLENGE_LEN  = 64

    JWT_SECRET_LEN = 256

    LONGPOLL_MAX = 30

    COLDWIRE_DATA_SEP byte  = 0
    COLDWIRE_LEN_OFFSET     = 3

    SQLITE_DB_NAME = "coldwire_database.sqlite"
    SQLI_DB_NAME   = "coldwire_database"
)
