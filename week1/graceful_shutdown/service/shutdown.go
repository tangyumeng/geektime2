package service

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"sync"
	"syscall"
	"time"

	"github.com/dubbogo/gost/log/logger"
)

// 典型的 Option 设计模式
type Option func(*App)

// ShutdownCallback 采用 context.Context 来控制超时，而不是用 time.After 是因为
// - 超时本质上是使用这个回调的人控制的
// - 我们还希望用户知道，他的回调必须要在一定时间内处理完毕，而且他必须显式处理超时错误
type ShutdownCallback func(ctx context.Context)

// 你需要实现这个方法
func WithShutdownCallbacks(cbs ...ShutdownCallback) Option {
	return func(a *App) {
		a.cbs = cbs
	}
}

// 这里我已经预先定义好了各种可配置字段
type App struct {
	servers []*Server

	// 优雅退出整个超时时间，默认30秒
	shutdownTimeout time.Duration

	// 优雅退出时候等待处理已有请求时间，默认10秒钟
	waitTime time.Duration
	// 自定义回调超时时间，默认三秒钟
	cbTimeout time.Duration

	cbs []ShutdownCallback
}

// NewApp 创建 App 实例，注意设置默认值，同时使用这些选项
func NewApp(servers []*Server, opts ...Option) *App {
	a := App{
		servers:         servers,
		shutdownTimeout: 30,
		waitTime:        10,
		cbTimeout:       3,
		cbs:             nil,
	}

	for _, o := range opts {
		o(&a)
	}
	return &a
}

// StartAndServe 你主要要实现这个方法
func (app *App) StartAndServe() {
	for _, s := range app.servers {
		srv := s
		go func() {
			if err := srv.Start(); err != nil {
				if err == http.ErrServerClosed {
					log.Printf("服务器%s已关闭", srv.name)
				} else {
					log.Printf("服务器%s异常退出", srv.name)
				}

			}
		}()
	}
	// 从这里开始优雅退出监听系统信号，强制退出以及超时强制退出。
	// 优雅退出的具体步骤在 shutdown 里面实现
	// 所以你需要在这里恰当的位置，调用 shutdown

	shutdownSignals := []os.Signal{
		os.Interrupt, os.Kill, syscall.SIGKILL, syscall.SIGSTOP,
		syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGILL, syscall.SIGTRAP,
		syscall.SIGABRT, syscall.SIGSYS, syscall.SIGTERM,
	}
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, shutdownSignals...)

	dumpHeapShutdownSignals := []os.Signal{
		syscall.SIGQUIT, syscall.SIGILL,
		syscall.SIGTRAP, syscall.SIGABRT, syscall.SIGSYS,
	}

	go func() {
		select {
		case sig := <-signals:
			logger.Infof("get signal %s, applicationConfig will shutdown.", sig)
			// gracefulShutdownOnce.Do(func() {
			time.AfterFunc(app.shutdownTimeout, func() {
				logger.Warn("Shutdown gracefully timeout, applicationConfig will shutdown immediately. ")
				os.Exit(0)
			})
			app.shutdown()
			// those signals' original behavior is exit with dump ths stack, so we try to keep the behavior
			for _, dumpSignal := range dumpHeapShutdownSignals {
				if sig == dumpSignal {
					debug.WriteHeapDump(os.Stdout.Fd())
				}
			}
			os.Exit(0)
		}
	}()
}

// shutdown 你要设计这里面的执行步骤。
func (app *App) shutdown() {
	log.Println("开始关闭应用，停止接收新请求")
	// 你需要在这里让所有的 server 拒绝新请求
	for _, s := range app.servers {
		(*(s.srv)).reject = true
	}

	log.Println("等待正在执行请求完结")
	// 在这里等待一段时间
	time.Sleep(app.waitTime)

	log.Println("开始关闭服务器")
	// 并发关闭服务器，同时要注意协调所有的 server 都关闭之后才能步入下一个阶段

	wg := sync.WaitGroup{}

	for i, s := range app.servers {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			fmt.Printf("stop %v server", idx)
			s.stop()
		}(i)
	}

	wg.Wait()

	log.Println("开始执行自定义回调")
	// 并发执行回调，要注意协调所有的回调都执行完才会步入下一个阶段

	for i, cb := range app.cbs {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			fmt.Println("call cb in cbs at index", idx)
			cb(context.Background()) //TODO: check where context from
		}(i)
	}

	wg.Wait()

	// 释放资源
	log.Println("开始释放资源")
	app.close()
}

func (app *App) close() {
	// 在这里释放掉一些可能的资源
	time.Sleep(time.Second)
	log.Println("应用关闭")
}

type Server struct {
	srv  *http.Server
	name string
	mux  *serverMux
}

// serverMux 既可以看做是装饰器模式，也可以看做委托模式
type serverMux struct {
	reject bool
	*http.ServeMux
}

func (s *serverMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.reject {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("服务已关闭"))
		return
	}
	s.ServeMux.ServeHTTP(w, r)
}

func NewServer(name string, addr string) *Server {
	mux := &serverMux{ServeMux: http.NewServeMux()}
	return &Server{
		name: name,
		mux:  mux,
		srv: &http.Server{
			Addr:    addr,
			Handler: mux,
		},
	}
}

func (s *Server) Handle(pattern string, handler http.Handler) {
	s.mux.Handle(pattern, handler)
}

func (s *Server) Start() error {
	return s.srv.ListenAndServe()
}

func (s *Server) rejectReq() {
	s.mux.reject = true
}

func (s *Server) stop() error {
	log.Printf("服务器%s关闭中", s.name)
	return s.srv.Shutdown(context.Background())
}