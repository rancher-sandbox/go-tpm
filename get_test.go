package tpm_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/rancher-sandbox/go-tpm"
)

func writeRead(conn *websocket.Conn, input []byte) ([]byte, error) {
	writer, err := conn.NextWriter(websocket.BinaryMessage)
	if err != nil {
		return nil, err
	}

	if _, err := writer.Write(input); err != nil {
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	_, reader, err := conn.NextReader()
	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(reader)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Mimics a WS server which accepts TPM Bearer token
func WSServer(ctx context.Context) {
	s := http.Server{
		Addr:         ":8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	m := http.NewServeMux()
	m.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		conn, _ := upgrader.Upgrade(w, r, nil) // error ignored for sake of simplicity

		for {

			token := r.Header.Get("Authorization")
			ek, at, err := GetAttestationData(token)
			if err != nil {
				fmt.Println("error", err.Error())
				return
			}

			secret, challenge, err := GenerateChallenge(ek, at)
			if err != nil {
				fmt.Println("error", err.Error())
				return
			}

			resp, _ := writeRead(conn, challenge)

			if err := ValidateChallenge(secret, resp); err != nil {
				fmt.Println("error validating challenge", err.Error())
				return
			}

			writer, _ := conn.NextWriter(websocket.BinaryMessage)
			json.NewEncoder(writer).Encode(map[string]string{"foo": "bar"})
		}
	})

	s.Handler = m

	go s.ListenAndServe()
	go func() {
		<-ctx.Done()
		s.Shutdown(ctx)
	}()
}

var _ = Describe("GET", func() {
	Context("challenges", func() {
		It("fails for permissions", func() {
			_, err := Get("http://localhost:8080/test")
			Expect(err).To(HaveOccurred())
		})
		It("gets pubhash", func() {

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			WSServer(ctx)

			msg, err := Get("http://localhost:8080/test", Emulated, WithSeed(1))
			result := map[string]interface{}{}
			json.Unmarshal(msg, &result)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(map[string]interface{}{"foo": "bar"}))
		})
	})
})
