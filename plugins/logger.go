package plugins

import (
	"github.com/dekinsq/qf"
	"log"
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
