package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/mail"
	"os"
	"read_creds"
	"strings"
	"time"

	"code.google.com/p/go-imap/go1/imap"
)

const (
	imap_server string = "imap.example.com:993"
)

var (
	client    *imap.Client
	cmd, cmd2 *imap.Command
	rsp       *imap.Response
	rsp2      *imap.Response
	err       error
	subject   string
	filename  string = "results.xml"
)

func main() {

	fp, err := os.Create(filename)
	defer fp.Close()

	reader := bufio.NewReader(os.Stdin)

	// Connect to the server
	client, err = imap.DialTLS(imap_server, nil)
	if err != nil {
		fmt.Println("Error on connecting to the imap server")
		fmt.Println(err.Error())
		return
	}

	// Remember to log out and close the connection when finished
	defer client.Logout(30 * time.Second)

	// Print server greeting (first response in the unilateral server data queue)
	fmt.Println("Server says hello:", client.Data[0].Info)
	client.Data = nil

	fmt.Println("Reading credentials")
	creds := strings.Split(read_creds.ReadCreds(), ",")

	// Authenticate
	if client.State() == imap.Login {
		client.Login(creds[0], creds[1])
	}

	// Open a mailbox (synchronous command - no need for imap.Wait)
	client.Select("INBOX", true)

	// Fetch the headers of the 20 most recent messages
	set, _ := imap.NewSeqSet("")
	if client.Mailbox.Messages >= 20 {
		set.AddRange(client.Mailbox.Messages-19, client.Mailbox.Messages)
	} else {
		set.Add("1:*")
	}

	cmd, err = client.Fetch(set, "RFC822.HEADER", "RFC822.TEXT")
	if err != nil {
		fmt.Printf("Error on header fetch, err: %s\n", err)
		return
	}

	// Process responses while the command is running
	for cmd.InProgress() {
		// Wait for the next response (no timeout)
		client.Recv(-1)

		// Process command data
		for _, rsp = range cmd.Data {
			header := imap.AsBytes(rsp.MessageInfo().Attrs["RFC822.HEADER"])
			if msg, _ := mail.ReadMessage(bytes.NewReader(header)); msg != nil {
				subject = msg.Header.Get("Subject")
				if "Daily Employee Transfer Notice" == subject {
					fmt.Println("Subject: ", subject)
					mediaType, params, err := mime.ParseMediaType(
						msg.Header.Get("Content-Type"))
					if err != nil {
						fmt.Println("Error locating Mime type")
						return
					}
					if strings.HasPrefix(mediaType, "multipart/") {
						body := strings.NewReader(imap.AsString(rsp.MessageInfo().Attrs["RFC822.TEXT"]))
						mr := multipart.NewReader(body, params["boundary"])
						for p, err := mr.NextPart(); err != io.EOF; p, err = mr.NextPart() {
							if err != nil {
								fmt.Println("Error reading message body")
								fmt.Println(err)
								return
							}

							slurp, err := ioutil.ReadAll(p)
							if err != nil {
								fmt.Println("Error slurping message body")
								fmt.Println(err)
								return
							}
							contType := p.Header.Get("Content-Type")
							if strings.HasPrefix(contType, "application/octet-stream; name=") && strings.HasSuffix(contType, `.xml"`) {
								fmt.Println("Content-Type: ", contType)
								fmt.Printf("Slurp len: %d\n", len(slurp))
								fmt.Println("Press enter to continue")
								text, _ := reader.ReadString('\n')
								fmt.Println(text)
								slurp_decoded, err := base64.StdEncoding.DecodeString(string(slurp))
								if err != nil {
									fmt.Println("Error decoding base64 content: ", err)
									// return
								} else {
									fp.WriteString(string(slurp_decoded))
									fp.WriteString("\n----------------------------------------------\n")
									// flag: \Deleted added.
									cmd, err = client.Store(another_set, "item", value)
								}
							}
						}
					}
				}
			}
		}
		cmd.Data = nil
		client.Data = nil

		// Process unilateral server data
		for _, rsp = range client.Data {
			fmt.Println("Server data:", rsp)
		}
		client.Data = nil
	}
	cmd, err = client.Expunge()

	// Check command completion status
	if rsp, err := cmd.Result(imap.OK); err != nil {
		if err == imap.ErrAborted {
			fmt.Println("Fetch command aborted")
		} else {
			fmt.Println("Fetch error:", rsp.Info)
		}
	}

}
