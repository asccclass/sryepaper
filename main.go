/*

*/
package main

import(
   "os"
   "fmt"
   "net"
   "time"
   "runtime"
   "context"
   "syscall"
   "os/signal"
   "github.com/joho/godotenv"
)

type Sryepaper struct {
   path		string		`json:"path"`
   port		string		`json:"port"` 
}

func(epaper *Sryepaper) handleRequest(conn net.Conn)  {
   remoteAddr := conn.RemoteAddr().String()
   fmt.Println("Client connected from:" + remoteAddr)
   buf := make([]byte, 2048)
   n, err := conn.Read(buf)
   if err != nil {
      fmt.Println(err.Error())
      return
   }
   fileName := string(buf[:n])
   conn.Write([]byte("ok"))
   // Create file
   f, err := os.Create(epaper.path + "/" + fileName)
   if err != nil {
      fmt.Println(err.Error())
      return
   }
   for {
      buf := make([]byte, 1024)
      n, _ := conn.Read(buf)
      if string(buf[:n]) == "finish" {
         runtime.Goexit()
      }
      f.Write(buf[:n])
   }
   defer f.Close()
   defer conn.Close()
}

func offLine(ctx context.Context) {
   fmt.Println("\nSystem closing")
}

func main() {
   if err := godotenv.Load("envfile"); err != nil {
      fmt.Println("no envfile")
      return
   }
   port := os.Getenv("PORT")
   if port == "" {
      fmt.Println("PORT is not set")
      return
   } 
   path := os.Getenv("SaveFileDir")
   if path == "" {
      fmt.Println("SaveFileDir is not set in envfile")
      return
   }
   epr := &Sryepaper {
      port: port,
      path: path,
   }
   l, err := net.Listen("tcp", ":" + epr.port)
   if err != nil {
      fmt.Println(err.Error())
      return
   }
   defer l.Close()
   // deal with exit
   interrupt := make(chan os.Signal)
   signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
   go func() {
      <-interrupt
      ctx, cancel := context.WithTimeout(context.Background(), time.Second * 15)
      defer cancel()
      offLine(ctx)   // exit before doing this
      return
   }()

   for {
      conn, err := l.Accept()
      if err != nil {
         fmt.Println(err.Error())
         return
      }
      go epr.handleRequest(conn)
   }
}
