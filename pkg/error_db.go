package pkg

import (
    "errors"
    "fmt"
    "net"

    "github.com/lib/pq"
    "github.com/jackc/pgx/v5/pgconn"
    "gorm.io/gorm"
)

// PostgreSQL Error Codes
// Referensi: https://www.postgresql.org/docs/current/errcodes-appendix.html
const (
    // Class 08 — Connection Exception
    PgErrConnectionException         = "08000"
    PgErrConnectionDoesNotExist      = "08003"
    PgErrConnectionFailure           = "08006"

    // Class 23 — Integrity Constraint Violation
    PgErrIntegrityConstraintViolation = "23000"
    PgErrRestrictViolation            = "23001"
    PgErrNotNullViolation             = "23502"
    PgErrForeignKeyViolation          = "23503"
    PgErrUniqueViolation              = "23505" // duplicate / already exist
    PgErrCheckViolation               = "23514"
    PgErrExclusionViolation           = "23P01"

    // Class 22 — Data Exception
    PgErrDataException                = "22000"
    PgErrNumericValueOutOfRange       = "22003"
    PgErrInvalidTextRepresentation    = "22P02" // invalid UUID, enum, etc.
    PgErrStringDataRightTruncation    = "22001"
    PgErrDivisionByZero               = "22012"

    // Class 42 — Syntax Error or Access Rule Violation
    PgErrUndefinedTable               = "42P01"
    PgErrUndefinedColumn              = "42703"
    PgErrInsufficientPrivilege        = "42501"

    // Class 53 — Insufficient Resources
    PgErrTooManyConnections           = "53300"
    PgErrDiskFull                     = "53100"
    PgErrOutOfMemory                  = "53200"

    // Class 40 — Transaction Rollback
    PgErrTransactionRollback          = "40000"
    PgErrSerializationFailure         = "40001"
    PgErrDeadlockDetected             = "40P01"

    // Class 57 — Operator Intervention
    PgErrQueryCanceled                = "57014"
    PgErrAdminShutdown                = "57P01"
    PgErrCrashShutdown                = "57P02"

    // Class 55 — Object Not In Prerequisite State
    PgErrObjectNotInPrerequisiteState = "55000"
    PgErrLockNotAvailable             = "55P03"
)

// Domain errors — gunakan ini di layer atas (service/handler)
var (
    ErrAlreadyExists         = errors.New("record already exists!")
    ErrNotFound              = errors.New("record not found!")
    ErrForeignKeyViolation   = errors.New("related record not found or still referenced!")
    ErrNotNullViolation      = errors.New("required field is missing!")
    ErrCheckViolation        = errors.New("value does not satisfy constraint!")
    ErrDataTooLong           = errors.New("data exceeds maximum length!")
    ErrInvalidDataFormat     = errors.New("invalid data format!")
    ErrDeadlock              = errors.New("deadlock detected, please retry!")
    ErrSerializationFailure  = errors.New("transaction conflict, please retry!")
    ErrConnectionFailed      = errors.New("database connection failed!")
    ErrTooManyConnections    = errors.New("database is overloaded, too many connections!")
    ErrQueryCanceled         = errors.New("query was canceled due to timeout!")
    ErrInsufficientPrivilege = errors.New("insufficient database privilege!")
    ErrUndefinedTable        = errors.New("table does not exist!")
    ErrUndefinedColumn       = errors.New("column does not exist!")
    ErrInternal              = errors.New("internal database error!")
    ErrEmailAlreadyUsed      = errors.New("email already used, please resend to your email for verify your account")
    ErrEmailSentAgain        = errors.New("email was already registered but not verified. a new verification link has been sent to your email.")
    ErrEmailAlreadyVerified  = errors.New("account already verified, please sign in to your account!")
    ErrUsernameAlreadyUsed   = errors.New("username already used")
    ErrEmailNotVerified      = errors.New("email not verified")
    ErrInvalidCredentials    = errors.New("invalid email or password")
    ErrInvalidPassword       = errors.New("invalid password")
    ErrForbidden             = errors.New("access denied")
    ErrInvalidOrExpiredToken = errors.New("invalid or expired token")
    ErrDeletePrimaryAddress  = errors.New("primary address cannot be deleted")
)

// HandleDBError memetakan raw DB error ke domain error yang bersih.
// Selalu kembalikan error ini ke caller; jangan expose *pq.Error ke luar repository.
func HandleDBError(err error) error {
    if err == nil {
        return nil
    }

    // --- GORM built-in errors ---
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return ErrNotFound
    }
    if errors.Is(err, gorm.ErrDuplicatedKey) {
        return ErrAlreadyExists
    }
    if errors.Is(err, gorm.ErrForeignKeyViolated) {
        return ErrForeignKeyViolation
    }
    if errors.Is(err, gorm.ErrCheckConstraintViolated) {
        return ErrCheckViolation
    }

    // --- pgx driver: *pgconn.PgError ---
    var pgErr *pgconn.PgError
    if errors.As(err, &pgErr) {
        return handlePGError(pgErr)
    }

    // --- lib/pq driver: *pq.Error (fallback) ---
    var pqErr *pq.Error
    if errors.As(err, &pqErr) {
        return handlePQError(pqErr)
    }

    // --- Network errors ---
    var netErr *net.OpError
    if errors.As(err, &netErr) {
        return fmt.Errorf("%w: %s", ErrConnectionFailed, netErr.Op)
    }

    return fmt.Errorf("%w: %s", ErrInternal, err.Error())
}

func handlePGError(err *pgconn.PgError) error {
    switch err.Code {
    case PgErrUniqueViolation:
        if err.ConstraintName != "" {
            return fmt.Errorf("%w: constraint=%s", ErrAlreadyExists, err.ConstraintName)
        }
        return ErrAlreadyExists

    case PgErrForeignKeyViolation:
        if err.ConstraintName != "" {
            return fmt.Errorf("%w: constraint=%s", ErrForeignKeyViolation, err.ConstraintName)
        }
        return ErrForeignKeyViolation

    case PgErrNotNullViolation:
        if err.ColumnName != "" {
            return fmt.Errorf("%w: column=%s", ErrNotNullViolation, err.ColumnName)
        }
        return ErrNotNullViolation

    case PgErrCheckViolation:
        if err.ConstraintName != "" {
            return fmt.Errorf("%w: constraint=%s", ErrCheckViolation, err.ConstraintName)
        }
        return ErrCheckViolation

    case PgErrDeadlockDetected:
        return ErrDeadlock

    case PgErrSerializationFailure:
        return ErrSerializationFailure

    case PgErrTooManyConnections:
        return ErrTooManyConnections

    case PgErrQueryCanceled:
        return ErrQueryCanceled

    case PgErrInsufficientPrivilege:
        return ErrInsufficientPrivilege

    // ... mirror the rest of handlePQError cases

    default:
        return fmt.Errorf("%w: pg_code=%s msg=%s", ErrInternal, err.Code, err.Message)
    }
}

func handlePQError(err *pq.Error) error {
    code := string(err.Code)

    switch code {

    // ── Unique / duplicate ──────────────────────────────────────────────
    case PgErrUniqueViolation:
        if err.Constraint != "" {
            return fmt.Errorf("%w: constraint=%s", ErrAlreadyExists, err.Constraint)
        }
        return ErrAlreadyExists

    // ── Foreign key ─────────────────────────────────────────────────────
    case PgErrForeignKeyViolation:
        if err.Constraint != "" {
            return fmt.Errorf("%w: constraint=%s", ErrForeignKeyViolation, err.Constraint)
        }
        return ErrForeignKeyViolation

    // ── NOT NULL ────────────────────────────────────────────────────────
    case PgErrNotNullViolation:
        if err.Column != "" {
            return fmt.Errorf("%w: column=%s", ErrNotNullViolation, err.Column)
        }
        return ErrNotNullViolation

    // ── CHECK constraint ────────────────────────────────────────────────
    case PgErrCheckViolation:
        if err.Constraint != "" {
            return fmt.Errorf("%w: constraint=%s", ErrCheckViolation, err.Constraint)
        }
        return ErrCheckViolation

    // ── Exclusion constraint ────────────────────────────────────────────
    case PgErrExclusionViolation:
        return fmt.Errorf("%w: exclusion constraint=%s", ErrCheckViolation, err.Constraint)

    // ── Data format / truncation ────────────────────────────────────────
    case PgErrStringDataRightTruncation:
        return fmt.Errorf("%w: column=%s", ErrDataTooLong, err.Column)
    case PgErrInvalidTextRepresentation, PgErrDataException:
        return fmt.Errorf("%w: %s", ErrInvalidDataFormat, err.Detail)
    case PgErrNumericValueOutOfRange:
        return fmt.Errorf("%w: numeric out of range column=%s", ErrInvalidDataFormat, err.Column)
    case PgErrDivisionByZero:
        return fmt.Errorf("%w: division by zero", ErrInvalidDataFormat)

    // ── Transaction / deadlock ──────────────────────────────────────────
    case PgErrDeadlockDetected:
        return ErrDeadlock
    case PgErrSerializationFailure:
        return ErrSerializationFailure
    case PgErrTransactionRollback:
        return fmt.Errorf("%w: transaction was rolled back", ErrInternal)

    // ── Connection ──────────────────────────────────────────────────────
    case PgErrConnectionException,
        PgErrConnectionDoesNotExist,
        PgErrConnectionFailure:
        return fmt.Errorf("%w: %s", ErrConnectionFailed, err.Message)

    // ── Resources ───────────────────────────────────────────────────────
    case PgErrTooManyConnections:
        return ErrTooManyConnections
    case PgErrDiskFull:
        return fmt.Errorf("%w: disk full", ErrInternal)
    case PgErrOutOfMemory:
        return fmt.Errorf("%w: out of memory", ErrInternal)

    // ── Operator intervention ───────────────────────────────────────────
    case PgErrQueryCanceled:
        return ErrQueryCanceled
    case PgErrAdminShutdown, PgErrCrashShutdown:
        return fmt.Errorf("%w: server is shutting down", ErrConnectionFailed)

    // ── Schema / privilege ──────────────────────────────────────────────
    case PgErrInsufficientPrivilege:
        return ErrInsufficientPrivilege
    case PgErrUndefinedTable:
        return fmt.Errorf("%w: %s", ErrUndefinedTable, err.Message)
    case PgErrUndefinedColumn:
        return fmt.Errorf("%w: %s", ErrUndefinedColumn, err.Message)

    // ── Lock ────────────────────────────────────────────────────────────
    case PgErrLockNotAvailable:
        return fmt.Errorf("%w: lock not available, try again", ErrDeadlock)

    default:
        return fmt.Errorf("%w: pg_code=%s msg=%s", ErrInternal, code, err.Message)
    }
}