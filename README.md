# secure-golang-web-app-demo
Demo app to demonstrate security mechanism implementations 

API based Web App using golang gorilla mux

Following are the security mechanisms implemented in the app

- https connection
- JWT authentication mechanism  
- Secret (jwt key) values are stored in env variables 
- JWT signing method is verified 
- Token cookie is on only my domain localhsot 
- Token expires after a period of time
- Token expires after logout
- Token is stored using HttpOnly cookie
- Password are hashed and salted
- Server side input validation 
- Password complexity checks
- CSRF protection 
- Logging events to syslog/logfile
- Cookie values expires and deleted from browser
- Autocomplete="off" for html sensitive input box
- Rate limit protection
- Disabling Caching  
- .gitignore keys, .env, and database

## Discussion

There are assumptions made for the nature of application.

- User is entering data in the form which may belong to him or any other person. The form allow to enter email and name even they are not matching the logged user info.
- Secret values are stored in environment variables, However for complex envirnment always recomend to use secret manager such as Vault or AWS Secrets Manager
- HTTPs uses self signed certificates just to stress on the importance of encrypted traffic, however certs should be always generated from CA. Let's encrypt is a good option for such demo. Implementation will not change for both CA signed or self-signed certs.
- Tokens expire after a period of time, 15 min in this app. There are services where session last for long time such as Twitter, Github. Gmail. The assumption is that the user should logout if he wants to terminate the session. In this app both logout and expiry time are implemented. Based on the criticality of the service, token expiry time is set. Redis has the feature to delete record after a period of time. 
- Blacklisted tokens due to logout functionality can be deleted after they expiry using database stored procedures. both token and expiry timestamp are stored in the same table.  
- For sake of making the app setup easy, Sqlite database is used, there are no credentials to access the database since it's stored locally on the server, but for remote databases such as mysql, postgress ...etc. Database should be secured with strong password.
- Security analysts and IR keep their jobs because of logs. The app implemented simple logging mechanism for major events such as user created, login success, login failure log out, message sent ..etc. It's always recommend to support logging to facilitate sending the logs to SIEM such as Splunk or ELK
- Rate limit is implemented on login page only as a demo, it offers protection against bruteforcing and also protect the server utilization. Locking down the user sending too many request is not implemented, it can be done using WAF. The current implementation does not limit by ip address since it's a localhost demo. It protects the overall function from being called too many times within time frame. The same can be done using  username if the purpose is not security rather API billing.
- Extra client side input validation checks could be added to improve user's experience however, they are ignored since they have been taken care of at the server side for security purpose
- Bad certificate notice will appear because the certificate is not fixed for ip, you should ignore this message, for production app always CA signed certificates should be used


**Setup instructions are for Linux systems**. Golang 1.15

Follow the steps blow and execute the given commands by copy paste to your terminal


First make sure you have go1.15+ installed

```sh
go version 
```

you should see something like

```sh
go version go1.15 linux/amd64
```


## Clone the repo

```sh
cd $HOME/go
go get github.com/alwashali/secure-golang-web-app-demo
```

## Generate ssl certificates

```sh
mkdir keys
openssl genrsa -out keys/server.key 2048
openssl req -new -x509 -sha256 -key keys/server.key -out keys/server.crt -days 3650 
```

## Create .env file

Create .env file which will be used for loading the keys to OS environment variables. You should put your 32 character long keys inside .env file. 
```sh 
echo "KEY={replace with your 32 character KEY}" > .env 
echo "CSRFKEY={replace with your  32 character CSRFKEY key}" >> .env
```

## Run

```sh
go run main.go
```

Once you are in enjoying the cat video, then create a new user by visiting https://localhost:8080/signup

