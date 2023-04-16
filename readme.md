# mailer

unit file format for configuration and mail setup

```/etc/example.mail```

```shell [gmail]
user=example@gmail.com
pass=
smtp=smtp.gmail.com

[admin]
mail=admin@gmail.com
```

Simple use case

```golang

	environ := env.NewEnv()

	// configure mail and store unit file path for later reference
	mailer := mail.NewMail( 
		env.Dir(environ.Etc, "example.mail"),
		"gmail",
	)

// send an admin alert using the stored unit file
	mailer.Alert().Send( // send an admin alert
		"admin",              // to; process [section] from the saved unit file
		nil,                  // subject
		"something happened", // message
	)



```