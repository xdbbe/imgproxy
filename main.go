package main

import (
	"bufio"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber"
	"github.com/gofiber/fiber/middleware"
	"github.com/patrickmn/go-cache"
)

var ca = cache.New(5*time.Minute, 10*time.Minute)

func main() {
	//Memory Profiler, disable in prod
	//defer profile.Start(profile.MemProfile).Stop()

	app := fiber.New(&fiber.Settings{
		//enable if there is enough ram (min 32-64 gig)
		//Prefork:          true,
		DisableKeepalive: true,
	})

	//SSL, enable if standalone with own ip
	// cer, err := tls.LoadX509KeyPair("server.crt", "server.key")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// config := &tls.Config{Certificates: []tls.Certificate{cer}}

	app.Use(middleware.Recover())
	//Logger
	f, _ := os.Create("imgproxy.log")
	defer f.Close()
	w := bufio.NewWriter(f)
	go func() {
		for {
			w.Flush()
			time.Sleep(1 * time.Second)
		}
	}()
	app.Use(middleware.Logger(w))
	//Main Logic
	app.Get("/img/:hash.:extension", func(c *fiber.Ctx) {

		if !checkCachePost(c) {
			extension := c.Params("extension")
			hash := c.Params("hash")

			resp, success := conv(hash, hash+"."+extension, c)

			if success == true {
				c.Set("content-type", "image/"+extension)
				ca.Set(strings.Join([]string{hash, extension, c.FormValue("width"), c.FormValue("height")}, ";"), resp, cache.DefaultExpiration)
			} else {
				c.Set("content-type", "application/xml")
			}
			if resp != nil {
				c.Send(resp)
			}
			c.Set("X-Powered-By", "xdb-imgproxy")
		}
	})

	app.Get("*", func(c *fiber.Ctx) {
		c.Status(404)
		c.Set("X-Powered-By", "xdb-imgproxy")
	})
	app.Listen(8080 /* , config */)
}
