package main

import (
  "fmt"
  "log"
  "encoding/json"
  "net/smtp"
  "net/mail"
  "encoding/base64"
  "strings"


  "github.com/streadway/amqp"
)

func failOnError(err error, msg string)  {
  if (err != nil) {
    log.Fatalf("%s: %s", msg, err)
    panic(fmt.Sprintf("%s: %s", msg, err))
  }
}

func sendEmail(email string) {
  log.Printf("Sending email to: %s", email)


  from := mail.Address{"Pit-Stop", "go.lang.demo.cs@gmail.com"}
  to := mail.Address{"Customer", email}
  title := "Car Reports"
  body := "Done, mate! Now you have Pit-Stop in the From :)"

  header := make(map[string]string)
  header["From"] = from.String()
  header["To"] = to.String()
  header["Subject"] = encodeRFC2047(title)
  header["MIME-Version"] = "1.0"
  header["Content-Type"] = "text/plain; charset=\"utf-8\""
  header["Content-Transfer-Encoding"] = "base64"

  message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + base64.StdEncoding.EncodeToString([]byte(body))

  log.Printf("Generated message content: " + message)
  // Set up authentication information.
  auth := smtp.PlainAuth(
      "",
      "go.lang.demo.cs@gmail.com",
      "g0.l@ng.dem0",
      "smtp.gmail.com",
  )
  // Connect to the server, authenticate, set the sender and recipient,
  // and send the email all in one step.
  err := smtp.SendMail(
      "smtp.gmail.com:587",
      auth,
      from.Address,
      []string{to.Address},
      []byte(message),
  )
  failOnError(err, "Failed to connect to email server")

  log.Printf("Sent email to: %s", email)
}

func encodeRFC2047(String string) string{
	// use mail's rfc2047 to encode any string
	addr := mail.Address{String, ""}
	return strings.Trim(addr.String(), " <>")
}

func main() {
  conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
  failOnError(err, "Failed to connect to RabbitMQ")
  defer conn.Close()

  ch, err := conn.Channel()
  failOnError(err, "Failed to open a channel")
  defer ch.Close()

  q, err := ch.QueueDeclare(
    "hello", // name
		false,   // durable
		false,   // delete when usused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
  )
  failOnError(err, "Failed to declare a queue")

  msgs, err := ch.Consume(
    q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
  )
  failOnError(err, "Failed to register a consumer")

  forever := make(chan bool)

  go func() {
    for d := range msgs {
        log.Printf("Received a message: %s", d.Body)

        var u user_struct
        err = json.Unmarshal([]byte(d.Body), &u)
        failOnError(err, "Failed to unmarshal message from the queue")

        sendEmail(u.Email)
    }
  }()

  log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
  <-forever
}
