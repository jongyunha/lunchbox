package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jongyunha/lunchbox/internal/config"
	"github.com/jongyunha/lunchbox/internal/system"
	"github.com/jongyunha/lunchbox/internal/web"
	"github.com/jongyunha/lunchbox/restaurants"
)

type monolith struct {
	*system.System
	modules []system.Module
}

func main() {
	if err := run(); err != nil {
		fmt.Printf("mallbots exitted abnormally: %s\n", err.Error())
		os.Exit(1)
	}
}

func run() (err error) {
	var cfg config.AppConfig
	cfg, err = config.InitConfig()
	if err != nil {
		return err
	}
	s, err := system.NewSystem(cfg)
	if err != nil {
		return err
	}
	m := monolith{
		System: s,
		modules: []system.Module{
			&restaurants.Module{},
		},
	}
	defer func(db *pgxpool.Pool) {
		db.Close()
	}(m.DB())

	if err = m.startupModules(); err != nil {
		return err
	}

	// Mount general web resources
	m.Mux().Mount("/", http.FileServer(http.FS(web.WebUI)))

	fmt.Println("started mallbots application")
	defer fmt.Println("stopped mallbots application")

	m.Waiter().Add(
		m.WaitForWeb,
		m.WaitForRPC,
		m.WaitForStream,
	)

	// go func() {
	// 	for {
	// 		var mem runtime.MemStats
	// 		runtime.ReadMemStats(&mem)
	// 		m.logger.Debug().Msgf("Alloc = %v  TotalAlloc = %v  Sys = %v  NumGC = %v", mem.Alloc/1024, mem.TotalAlloc/1024, mem.Sys/1024, mem.NumGC)
	// 		time.Sleep(10 * time.Second)
	// 	}
	// }()

	return m.Waiter().Wait()
}

func (m *monolith) startupModules() error {
	for _, module := range m.modules {
		ctx := m.Waiter().Context()
		if err := module.Startup(ctx, m); err != nil {
			return err
		}
	}

	return nil
}
