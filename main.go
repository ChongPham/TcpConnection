package main

import (
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// element in the queue
type Job struct {
	conn net.Conn
}

// represent thread in the pool
type Worker struct {
	id      int
	jobChan chan Job
	wg      *sync.WaitGroup
}

// the pool
type Pool struct {
	jobQueue chan Job
	workers  []*Worker
	wg       *sync.WaitGroup
}

// Create new worker/thread
func NewWorker(id int, jobChan chan Job, wg *sync.WaitGroup) *Worker {
	return &Worker{
		id:      id,
		jobChan: jobChan,
		wg:      wg,
	}
}

// Start worker/thread
func (w *Worker) Start() {
	go func() {
		for job := range w.jobChan {
			log.Printf("worker %d is handing job from %s", w.id, job.conn.RemoteAddr())
			handleConnection(job.conn)
			w.wg.Done()
		}
	}()
}

// Create new pool
func NewPool(numOfWorker int, limit int, wg *sync.WaitGroup) *Pool {
	return &Pool{
		jobQueue: make(chan Job, limit),
		workers:  make([]*Worker, numOfWorker),
		wg:       wg,
	}
}

// Start take jobs
func (p *Pool) Start() {
	for i := 0; i < len(p.workers); i++ {
		worker := NewWorker(i, p.jobQueue, p.wg)
		p.workers[i] = worker
		worker.Start()
	}
}

// push job to queue
func (p *Pool) AddJob(conn net.Conn) {
	p.jobQueue <- Job{conn: conn}
	p.wg.Add(1)
}

// Close Pool
func (p *Pool) Shutdown() {
	close(p.jobQueue) // báo cho worker dừng
	p.wg.Wait()       // chờ tất cả worker xong việc
}

func ReadCommand(conn net.Conn) (string, error) {
	var buf []byte = make([]byte, 512)
	n, err := conn.Read(buf[:])
	if err != nil {
		return "", err
	}

	return string(buf[:n]), nil
}

func WriteCommand(cmd string, conn net.Conn) error {
	_, err := conn.Write([]byte(cmd))
	if err != nil {
		return err
	}
	return nil
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	for {
		//read data from client
		cmd, err := ReadCommand(conn)
		if err != nil {
			if err == io.EOF {
				log.Println("Client closed connection:", conn.RemoteAddr())
				return
			}
			log.Println("Read error:", err)
			return
		}

		//process something
		time.Sleep(time.Second * 2)

		//rely to client
		var errRes = WriteCommand(cmd, conn)
		if errRes != nil {
			log.Print("err write:", errRes)
		}
	}
}

func main() {
	// tao tcp socket lang nghe o port 3000
	listener, err := net.Listen("tcp", ":3000")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	wg := &sync.WaitGroup{}
	pool := NewPool(4, 10, wg)
	pool.Start()

	// Bắt tín hiệu Ctrl+C hoặc kill
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Xử lí đóng poll
	go func() {
		<-sigChan
		log.Println("Shutting down server...")
		listener.Close() // ngắt Accept()
		pool.Shutdown()  // đóng pool
		os.Exit(0)
	}()

	for {
		// thiet lap dedicated connection
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}

		// Add job to jobChan
		pool.AddJob(conn)
	}
}
