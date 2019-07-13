# jeep

This application, though not complete, is a Golang application meant to run periodicallly (1/day) to 

1. Load email access credentials from an encrypted file.
2. Use those credentials to access an email account via IMAP, retrieve all new messages, looking for one matching a subject.
3. This message contains an XML attachment which represents an XML spreadsheet. This attachment is parsed and simple output is generated. 

