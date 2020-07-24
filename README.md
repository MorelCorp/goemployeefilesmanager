# goemployeefilesmanager

Employee File Manager for Google Drive in golang

Requirements

- go latest version (1.14.4 at the time of this writing)
- credentials.json (google API credentials - see FAQ)
- config.json ()

FAQ

How to get credentials.json
You need the file credentials.json in the launch folder which you can find by enabling google drive API as shown in this google example
<https://developers.google.com/drive/api/v3/quickstart/go>

    You also need to enable the spreadsheet API as specified in
    <https://developers.google.com/sheets/api/quickstart/go>

go get -u google.golang.org/api/sheets/v4
go get -u golang.org/x/oauth2/google

credentials.json needs to be in the folder then you need to build the go module and run it in the CLI (it can't be run in VSCode for the initial OAuth handshake)

You might want to modify the constants in main.go to fit your organisation
DomainName = "[Your domain name eg. "google.com"]"
