package servicetools

import (
	"context"
	"fmt"
	"net"
	"port/utils"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

var service utils.Service
var workerLimit = make(chan struct{}, 200)

func checkPort(ip string, port int) bool {
	address := fmt.Sprintf("%s:%d", ip, port)
	conn, err := net.DialTimeout("tcp", address, 2*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func updateStatus(ctx context.Context, db *pgxpool.Pool, id int, online bool) {
	if online {
		_, _ = db.Exec(ctx,
			`UPDATE services SET online = $1, last_seen = NOW() WHERE id = $2`,
			true, id)
	} else {
		_, _ = db.Exec(ctx,
			`UPDATE services SET online = $1 WHERE id = $2`,
			false, id)
	}
}

func StartPortMonitor(ctx context.Context, pool *pgxpool.Pool) {
	fmt.Println("Port monitoring service started...")

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rows, err := pool.Query(ctx,
				`SELECT id, host(remote_ip) AS remote_ip, remote_port FROM services`)
			if err != nil {
				fmt.Println("DB error:", err)
				continue
			}
			for rows.Next() {
				var s utils.Service
				if err := rows.Scan(&s.ID, &s.RemoteIP, &s.RemotePort); err != nil {
					fmt.Println("Scan error:", err)
					continue
				}
				go func(s utils.Service) {
					workerLimit <- struct{}{}
					defer func() { <-workerLimit }()
					online := checkPort(s.RemoteIP, s.RemotePort)
					updateStatus(ctx, pool, s.ID, online)
				}(s)
			}
			rows.Close()
		case <-ctx.Done():
			fmt.Println("Port monitor stopped.")
			return
		}
	}
}
