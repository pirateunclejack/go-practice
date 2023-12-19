package handlers

import (
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
)

type EndToEndSuite struct {
	suite.Suite
}

func TestEndToEndSuite(t *testing.T) {
    suite.Run(t, new(EndToEndSuite))
}

func (s *EndToEndSuite) TestPostHandler() {
	c := http.Client{}
	r, _ := c.Get("http://localhost:1323/posts")
	s.Equal(http.StatusOK, r.StatusCode)
}

func (s *EndToEndSuite) TestPostNoResult() {
	c := http.Client{}
	r, _ := c.Get("http://localhost:1323/post/9923")
	s.Equal(http.StatusOK, r.StatusCode)
	b, _ := io.ReadAll(r.Body)
	s.JSONEq(`{"status":"ok","data":[]}`, string(b))
}
