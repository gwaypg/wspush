package main

import (
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"os/signal"
	"time"

	"github.com/gwaylib/cert"
	"github.com/gwaypg/wspush/module/etc"
	"github.com/gwaypg/wspush/module/wsnode"

	"github.com/gorilla/websocket"
	"github.com/gwaylib/errors"
	"github.com/gwaylib/eweb"
	"github.com/gwaylib/qsql"
	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var (
	e = eweb.Default()

	// Configure the upgrader
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			//不检查跨域
			return true
		},
	}
)

func init() {
	wsconn := NewWsConn()
	svc := NewService(wsconn)
	if err := svc.(*serviceImpl).loadCallbackCache(); err != nil {
		log.Warn(errors.As(err))
	}
	// TODO: replace to rpcx?
	rpc.RegisterName(wsnode.RpcName, svc)

	e.Debug = os.Getenv("GIN_MODE") != "release"
	// middle ware
	e.Use(middleware.Gzip())

	e.Any("/wsnode/conn", func(c echo.Context) error {
		req := c.Request()
		resp := c.Response()

		//log.Info("conn in", req.URL.String())
		// Upgrade initial GET request to a websocket
		ws, err := upgrader.Upgrade(resp, req, nil)
		if err != nil {
			log.Error(errors.As(err))
			resp.Header().Add("extensions", "系统升级中")
			return c.String(500, "系统升级中")
		}
		if err := wsconn.HandleConn(c, ws); err != nil {
			log.Warn(errors.As(err))
			return c.String(500, "系统升级中")
		}
		return nil
	})
}

func main() {

	// loading cert file
	keyFile := etc.Etc.String("cmd/wsnode", "https_tls_key")
	certFile := etc.Etc.String("cmd/wsnode", "https_tls_cert")
	if len(keyFile) > 0 {
		if err := cert.AddFileCert(os.ExpandEnv(keyFile), os.ExpandEnv(certFile)); err != nil {
			panic(err)
		}
		// auto reload file cert every day
		go func() {
			ticker := time.NewTicker(24 * time.Hour)
			for {
				<-ticker.C
				if err := cert.AddFileCert(os.ExpandEnv(keyFile), os.ExpandEnv(certFile)); err != nil {
					log.Warn(errors.As(err))
				}
			}
		}()
	} else {
		log.Info("cert file not not, using auto cert")
		cert.AddAutoCert("lib10", "wsnode")
	}

	// for rpc
	go func() {
		addr := etc.Etc.String("cmd/wsnode", "rpc_listen")
		log.Infof("Rpc listen: %s", addr)
		conn, err := net.Listen("tcp", addr)
		if err != nil {
			log.Exit(2, errors.As(err))
			return
		}
		defer func() {
			qsql.Close(conn)
			log.Exit(0, "rpc conn exit")
		}()
		rpc.Accept(conn)
	}()

	// for wss release
	go func() {
		httpsAddr := etc.Etc.String("cmd/wsnode", "https_listen")

		log.Infof("Https listen : %s", httpsAddr)
		if err := e.StartTLSConfig(httpsAddr, cert.GetTLSConfig()); err != nil {
			log.Exit(2, errors.As(err))
			return
		}
	}()

	// ws for debug
	go func() {
		httpAddr := etc.Etc.String("cmd/wsnode", "http_listen")
		log.Infof("Http listen : %s", httpAddr)
		if err := e.Start(httpAddr); err != nil {
			log.Exit(2, errors.As(err))
			return
		}
	}()

	// exit event
	fmt.Println("[ctrl+c to exit]")
	end := make(chan os.Signal, 2)
	signal.Notify(end, os.Interrupt, os.Kill)
	<-end
}
