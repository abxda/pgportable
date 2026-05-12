package main

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// DBNode representa una base de datos en la raíz del árbol.
type DBNode struct {
	Name    string `json:"name"`
	Owner   string `json:"owner"`
	Size    string `json:"size"`
	Current bool   `json:"current"`
}

type SchemaNode struct {
	Name string `json:"name"`
}

type TableNode struct {
	Name string `json:"name"`
	Type string `json:"type"` // BASE TABLE | VIEW | ...
	Rows int64  `json:"rows"` // estimación
}

// connectDB abre una conexión libpq a la BD indicada usando las creds del manager.
func (m *Manager) connectDB(ctx context.Context, db string) (*pgx.Conn, error) {
	if !m.IsRunning() {
		return nil, fmt.Errorf("PostgreSQL no está corriendo")
	}
	uri := fmt.Sprintf("postgres://%s:%s@127.0.0.1:%d/%s?sslmode=disable&connect_timeout=5",
		m.cfg.User, m.cfg.Password, m.cfg.Port, db)
	cctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	return pgx.Connect(cctx, uri)
}

// ListDatabases devuelve las BD no-template visibles para el usuario actual.
func (m *Manager) ListDatabases(ctx context.Context) ([]DBNode, error) {
	conn, err := m.connectDB(ctx, m.cfg.Database)
	if err != nil {
		return nil, err
	}
	defer conn.Close(ctx)

	rows, err := conn.Query(ctx, `
		SELECT d.datname,
		       pg_catalog.pg_get_userbyid(d.datdba) AS owner,
		       pg_size_pretty(pg_database_size(d.datname)) AS size
		FROM pg_database d
		WHERE d.datistemplate = false
		ORDER BY d.datname;
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []DBNode
	for rows.Next() {
		var n DBNode
		if err := rows.Scan(&n.Name, &n.Owner, &n.Size); err != nil {
			return nil, err
		}
		n.Current = (n.Name == m.cfg.Database)
		out = append(out, n)
	}
	return out, rows.Err()
}

// ListSchemas devuelve los esquemas de la BD indicada (excluye los internos).
func (m *Manager) ListSchemas(ctx context.Context, db string) ([]SchemaNode, error) {
	conn, err := m.connectDB(ctx, db)
	if err != nil {
		return nil, err
	}
	defer conn.Close(ctx)

	rows, err := conn.Query(ctx, `
		SELECT schema_name
		FROM information_schema.schemata
		WHERE schema_name NOT IN ('pg_catalog','information_schema','pg_toast')
		  AND schema_name NOT LIKE 'pg_temp_%'
		  AND schema_name NOT LIKE 'pg_toast_temp_%'
		ORDER BY schema_name;
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []SchemaNode
	for rows.Next() {
		var n SchemaNode
		if err := rows.Scan(&n.Name); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

// ListTables devuelve tablas y vistas del esquema con estimación de filas.
func (m *Manager) ListTables(ctx context.Context, db, schema string) ([]TableNode, error) {
	conn, err := m.connectDB(ctx, db)
	if err != nil {
		return nil, err
	}
	defer conn.Close(ctx)

	rows, err := conn.Query(ctx, `
		SELECT t.table_name,
		       t.table_type,
		       COALESCE(c.reltuples, 0)::bigint AS row_estimate
		FROM information_schema.tables t
		LEFT JOIN pg_namespace n ON n.nspname = t.table_schema
		LEFT JOIN pg_class     c ON c.relname = t.table_name AND c.relnamespace = n.oid
		WHERE t.table_schema = $1
		ORDER BY t.table_name;
	`, schema)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []TableNode
	for rows.Next() {
		var n TableNode
		if err := rows.Scan(&n.Name, &n.Type, &n.Rows); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}
