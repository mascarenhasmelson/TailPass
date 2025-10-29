// package main

// import (
// 	"context"
// 	"fmt"
// 	"os"
// 	"time"

// 	"github.com/jackc/pgtype"
// 	"github.com/jackc/pgx/v4/pgxpool"
// )

// const dbMaxConns = 5

// var connString = "postgres://admin:StrongPassword123@192.168.30.249:5432/tunnel_services"

// func main() {
// 	ctx := context.Background()
// 	config, err := pgxpool.ParseConfig(connString)
// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, "Unable to parse DATABASE URL: %v\n", err)
// 		os.Exit(1)
// 	}
// 	config.MaxConns = dbMaxConns

// 	pool, err := pgxpool.ConnectConfig(ctx, config)
// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
// 		os.Exit(1)
// 	}
// 	defer pool.Close()
// 	fmt.Println("Successfully connected to PostgreSQL!")

// 	rows, err := pool.Query(ctx, "SELECT id, service_name, local_ip, local_port, remote_ip, remote_port, pid, created_at FROM services")
// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, "Query failed: %v\n", err)
// 		return
// 	}
// 	defer rows.Close()

// 	for rows.Next() {
// 		var (
// 			id          int
// 			serviceName pgtype.Text
// 			localIP     pgtype.Text
// 			localPort   int
// 			remoteIP    pgtype.Text
// 			remotePort  int
// 			pid         pgtype.Int4
// 			createdAt   time.Time
// 		)

// 		err := rows.Scan(&id, &serviceName, &localIP, &localPort, &remoteIP, &remotePort, &pid, &createdAt)
// 		if err != nil {
// 			fmt.Fprintf(os.Stderr, "Row scan failed: %v\n", err)
// 			continue
// 		}

// 		// Use .String only if Valid, else show <NULL>
// 		svcName := "<NULL>"
// 		if serviceName.Status == pgtype.Present {
// 			svcName = serviceName.String
// 		}
// 		lIP := "<NULL>"
// 		if localIP.Status == pgtype.Present {
// 			lIP = localIP.String
// 		}
// 		rIP := "<NULL>"
// 		if remoteIP.Status == pgtype.Present {
// 			rIP = remoteIP.String
// 		}
// 		pidValue := "<NULL>"
// 		if pid.Status == pgtype.Present {
// 			pidValue = fmt.Sprintf("%d", pid.Int)
// 		}

// 		fmt.Printf("%d: %s | %s:%d -> %s:%d | PID=%s  | Created: %s\n",
// 			id, svcName, lIP, localPort, rIP, remotePort, pidValue, createdAt.Format(time.RFC3339))

// 	}

// 	if rows.Err() != nil {
// 		fmt.Fprintf(os.Stderr, "Rows iteration error: %v\n", rows.Err())
// 	}
// }
