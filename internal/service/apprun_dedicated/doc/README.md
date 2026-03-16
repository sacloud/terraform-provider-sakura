So an AppRun dedicated's cluster needs a dedicated service principal for its
own.  Technically speaking we can create one during integration tests.

However "being able to create a service principal" is a stronger priviledge
than "being able to run an AppRun".  It might be a bad idea to require
unnecessary strong priviledge every time.

Let's separate that part.  This directory includes a set of working terraform
configuration files that creates a dedicated project, and a service principal.


```console
$ echo "TF_WORKSPACE=example" > .env
$ mise plan
$ mise apply
```

Before running this make sure you have a usacloud profile, of the same name as
you set to `TF_WORKSPACE` ("example" in above snppet).
