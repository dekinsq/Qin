package plugins

import (
	"log"
	"qf"
	"time"
)

func Logger() qf.HandlerFunc {
	return func(c *qf.Context) {
		t := time.Now()
		c.Next()
		log.Printf("[%d] %s in %v",
			c.StatusCode,
			c.Req.RequestURI,
			time.Since(t))
	}
}
