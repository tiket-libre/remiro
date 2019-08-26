package remiro

import (
	"io"
	"net"
	"testing"

	"github.com/gomodule/redigo/redis"
	"github.com/rafaeljusto/redigomock"
	"github.com/stretchr/testify/assert"
)

func Test_redisHandler_HandleGET(t *testing.T) {

	t.Run(`[Given] a key is available in "destination" 
		    [When] a GET request for the key is received
		    [Then] GET and return the key value from "destination"`, func(t *testing.T) {

		mockConn := redigomock.NewConn()
		handler := &redisHandler{
			destinationPool: &redis.Pool{
				Dial: func() (redis.Conn, error) {
					return mockConn, nil
				},
			},
		}

		mockCmd := mockConn.Command("GET", []byte("mykey")).Expect("hello")

		signal := make(chan error)
		go func() {
			if err := RunWithSignal(":1345", handler, signal); err != nil {
				t.Fatal(err)
			}
		}()

		done := make(chan bool)
		go func() {
			defer func() {
				done <- true
			}()

			err := <-signal
			if err != nil {
				t.Fatal(err)
			}

			conn, err := net.Dial("tcp", ":1345")
			if err != nil {
				t.Fatal(err)
			}
			defer conn.Close()

			_, err = io.WriteString(conn, "*2\r\n$3\r\nGET\r\n$5\r\nmykey\r\n")
			if err != nil {
				t.Fatal(err)
			}

			buf := make([]byte, 1024)
			n, err := conn.Read(buf)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, "$5\r\nhello\r\n", string(buf[:n]), "response should be equal to value")
			assert.True(t, mockCmd.Called, "redis should be called")
		}()

		<-done
	})

	t.Run(`[Given] a key is not available in "destination"
			 [And] the key is available in "source"
			 [And] deleteOnGet set to false
		    [When] a GET request for the key is received
		    [Then] GET and return the key value from "source"
		     [And] SET the value with the key to "destination"`, func(t *testing.T) {

		mockSrc := redigomock.NewConn()
		mockDst := redigomock.NewConn()
		handler := &redisHandler{
			sourcePool: &redis.Pool{
				Dial: func() (redis.Conn, error) {
					return mockSrc, nil
				},
			},
			destinationPool: &redis.Pool{
				Dial: func() (redis.Conn, error) {
					return mockDst, nil
				},
			},
			deleteOnGet: false,
		}

		mockDstGET := mockDst.Command("GET", []byte("mykey")).Expect(nil)
		mockDstSET := mockDst.Command("SET", []byte("mykey"), "hello").Expect("OK")
		mockSrcGET := mockSrc.Command("GET", []byte("mykey")).Expect("hello")

		signal := make(chan error)
		go func() {
			if err := RunWithSignal(":1346", handler, signal); err != nil {
				t.Fatal(err)
			}
		}()

		done := make(chan bool)
		go func() {
			defer func() {
				done <- true
			}()

			err := <-signal
			if err != nil {
				t.Fatal(err)
			}

			conn, err := net.Dial("tcp", ":1346")
			if err != nil {
				t.Fatal(err)
			}
			defer conn.Close()

			_, err = io.WriteString(conn, "*2\r\n$3\r\nGET\r\n$5\r\nmykey\r\n")
			if err != nil {
				t.Fatal(err)
			}

			buf := make([]byte, 1024)
			n, err := conn.Read(buf)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, "$5\r\nhello\r\n", string(buf[:n]), "response should be equal to value")
			assert.True(t, mockDstGET.Called, "destination redis GET command should be called")
			assert.True(t, mockDstSET.Called, "destination redis SET command should be called")
			assert.True(t, mockSrcGET.Called, "source redis GET command should be called")
		}()

		<-done
	})

	t.Run(`[Given] a key is not available in "destination"
			 [And] the key is available in "source"
			 [And] deleteOnGet set to true
		    [When] a GET request for the key is received
		    [Then] GET and return the key value from "source"
			 [And] SET the value with the key to "destination"
			 [And] DELETE the key from "source"`, func(t *testing.T) {

		mockSrc := redigomock.NewConn()
		mockDst := redigomock.NewConn()
		handler := &redisHandler{
			sourcePool: &redis.Pool{
				Dial: func() (redis.Conn, error) {
					return mockSrc, nil
				},
			},
			destinationPool: &redis.Pool{
				Dial: func() (redis.Conn, error) {
					return mockDst, nil
				},
			},
			deleteOnGet: true,
		}

		mockDstGET := mockDst.Command("GET", []byte("mykey")).Expect(nil)
		mockDstSET := mockDst.Command("SET", []byte("mykey"), "hello").Expect("OK")
		mockSrcGET := mockSrc.Command("GET", []byte("mykey")).Expect("hello")
		mockSrcDEL := mockSrc.Command("DEL", []byte("mykey")).Expect(1)

		signal := make(chan error)
		go func() {
			if err := RunWithSignal(":1347", handler, signal); err != nil {
				t.Fatal(err)
			}
		}()

		done := make(chan bool)
		go func() {
			defer func() {
				done <- true
			}()

			err := <-signal
			if err != nil {
				t.Fatal(err)
			}

			conn, err := net.Dial("tcp", ":1347")
			if err != nil {
				t.Fatal(err)
			}
			defer conn.Close()

			_, err = io.WriteString(conn, "*2\r\n$3\r\nGET\r\n$5\r\nmykey\r\n")
			if err != nil {
				t.Fatal(err)
			}

			buf := make([]byte, 1024)
			n, err := conn.Read(buf)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, "$5\r\nhello\r\n", string(buf[:n]), "response should be equal to value")
			assert.True(t, mockDstGET.Called, "destination redis GET command should be called")
			assert.True(t, mockDstSET.Called, "destination redis SET command should be called")
			assert.True(t, mockSrcGET.Called, "source redis GET command should be called")
			assert.True(t, mockSrcDEL.Called, "source redis DEL command should be called")
		}()

		<-done
	})

	t.Run(`[Given] a key is not available in "destination"
			 [And] the key is not available in "source"
		    [When] a GET request for the key is received
		    [Then] return nil message from "source"`, func(t *testing.T) {

		mockSrc := redigomock.NewConn()
		mockDst := redigomock.NewConn()
		handler := &redisHandler{
			sourcePool: &redis.Pool{
				Dial: func() (redis.Conn, error) {
					return mockSrc, nil
				},
			},
			destinationPool: &redis.Pool{
				Dial: func() (redis.Conn, error) {
					return mockDst, nil
				},
			},
		}

		mockDstGET := mockDst.Command("GET", []byte("mykey")).Expect(nil)
		mockSrcGET := mockSrc.Command("GET", []byte("mykey")).Expect(nil)

		signal := make(chan error)
		go func() {
			if err := RunWithSignal(":1348", handler, signal); err != nil {
				t.Fatal(err)
			}
		}()

		done := make(chan bool)
		go func() {
			defer func() {
				done <- true
			}()

			err := <-signal
			if err != nil {
				t.Fatal(err)
			}

			conn, err := net.Dial("tcp", ":1348")
			if err != nil {
				t.Fatal(err)
			}
			defer conn.Close()

			_, err = io.WriteString(conn, "*2\r\n$3\r\nGET\r\n$5\r\nmykey\r\n")
			if err != nil {
				t.Fatal(err)
			}

			buf := make([]byte, 1024)
			n, err := conn.Read(buf)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, "$-1\r\n", string(buf[:n]), "response should be equal to nil")
			assert.True(t, mockDstGET.Called, "destination redis GET command should be called")
			assert.True(t, mockSrcGET.Called, "source redis GET command should be called")
		}()

		<-done
	})
}

func Test_redisHandler_HandleSET(t *testing.T) {

	t.Run(`[Given] deleteOnSet set to false
			[When] a SET request for a key is received
			[Then] SET the key with the value to "destination"`, func(t *testing.T) {

	})

	t.Run(`[Given] deleteOnSet set to true
			[When] a SET request for a key is received
			[Then] SET the key with the value to "destination"
			 [And] DELETE the key from "source"`, func(t *testing.T) {

	})
}

func Test_redisHandler_HandlePING(t *testing.T) {

	t.Run(`[When] a PING request is received
		   [Then] return PONG`, func(t *testing.T) {

	})
}

func Test_redisHandler_HandleDefault(t *testing.T) {

	t.Run(`[When] any request except GET, SET, and PING is received
		   [Then] forward the request to "destination"
		    [And] return the response`, func(t *testing.T) {

	})
}