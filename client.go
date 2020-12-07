package pq

import (
	"context"
	"github.com/jackc/pgx/v4"
)

type Client interface {
	Executor
	Transactor
}

type Executor interface {
	// Exec executes sql. sql can be either a prepared statement name or an SQL string.
	// arguments should be referenced positionally from the sql string as $1, $2, etc.
	Exec(ctx context.Context, sql string, args ...interface{}) (result RowsAffected, err error)
	// Query executes sql with args. If there is an error the returned Rows will
	// be returned in an error state. So it is allowed to ignore the error returned
	// from Query and handle it in Rows.
	Query(ctx context.Context, sql string, args ...interface{}) (Rows, error)
	// QueryRow is a convenience wrapper over Query. Any error that occurs while
	// querying is deferred until calling Scan on the returned Row.
	QueryRow(ctx context.Context, sql string, args ...interface{}) Row
	// SendBatch sends all queued queries to the server at once. All queries are run in an implicit transaction unless
	// explicit transaction control statements are executed. The returned BatchResults must be closed before the connection
	// is used again.
	SendBatch(ctx context.Context, batch *pgx.Batch) BatchResults
}

type Transactor interface {
	Transaction(ctx context.Context, f func(context.Context, Executor) error) error
}

type RowsAffected interface {
	RowsAffected() int64
}

// Rows is the result set returned from *Client.Query. Rows must be closed before
// the Client can be used again. Rows are closed by explicitly calling Close(),
// calling Next() until it returns false, or when a fatal error occurs.
type Rows interface {
	// Close closes the rows, making the connection ready for use again. It is safe
	// to call Close after rows is already closed.
	Close()
	Err() error
	// Next prepares the next row for reading. It returns true if there is another
	// row and false if no more rows are available. It automatically closes rows
	// when all rows are read.
	Next() bool
	// Scan reads the values from the current row into dest values positionally.
	// dest can include pointers to core types, values implementing the Scanner
	// interface, []byte, and nil. []byte will skip the decoding process and directly
	// copy the raw bytes received from PostgreSQL. nil will skip the value entirely.
	Scan(dest ...interface{}) (err error)
}

// Scan works the same as (*Rows Scan) with the following exceptions. If no
// rows were found it returns ErrNoRows. If multiple rows are returned it
// ignores all but the first.
type Row interface {
	// Scan works the same as (*Rows Scan) with the following exceptions. If no
	// rows were found it returns ErrNoRows. If multiple rows are returned it
	// ignores all but the first.
	Scan(dest ...interface{}) (err error)
}

type BatchResults interface {
	// Exec reads the results from the next query in the batch as if the query has been sent with Conn.Exec.
	Exec() (RowsAffected, error)

	// Query reads the results from the next query in the batch as if the query has been sent with Conn.Query.
	Query() (Rows, error)

	// QueryRow reads the results from the next query in the batch as if the query has been sent with Conn.QueryRow.
	QueryRow() Row

	// Close closes the batch operation. This must be called before the underlying connection can be used again. Any error
	// that occurred during a batch operation may have made it impossible to resyncronize the connection with the server.
	// In this case the underlying connection will have been closed.
	Close() error
}
