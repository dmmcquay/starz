Starz
=============================

Starz uses github oauth to login and then does a simple query on a github 
account to return their public repos and how many stars each one has.

Installing From Source
==============================

Move the `starz` directory in the root of your `$GOPATH`.  Once there, you can 
install by running

```bash
$ go install starz/...
```

This will yield a binary called `starzd`. `starzd` comes with its own usage:


```bash
$ starzd run -h

run starzd with given options

Usage:
  app run [flags]

Flags:
  -a, --apitoken string       github apitoken
  -i, --clientId string       github oauth clientId
  -s, --clientSecret string   github oauth clientSecret
  -c, --cookieSecret string   cookieSecret
  -n, --host string           hostname
  -p, --port int              port (default -1)
```

`starzd` can be run by giving these flags values.  For example:

```bash
$ starzd run -i someClientIdFromGitHub -s someClientSecretFromGihub
```

is the most basic way to run the application.  You must get the clientid and 
clientsecret by creating a github oauth app. To register a github oauth app
and get a clientid and clientsecret, please visit [here](https://github.com/settings/applications/new)
and fill out the required information.

Github oauth also requires a github oauth call back route.  You can just default
this to:

```bash
http://127.0.0.1:$PORT/api/v0/github_oauth_cb/
```

unless you have an another ip you can use.

All the other flags are not required, but allow you to customize your experience.
For example, the apitoken, if not present, limits you to only 60 requests per
min.

The port, hostname, and cookieSecrets default to sane defaults (8000, 
localhost, and a random string).

Running a Docker Container
==============================

[Dockerhub](https://hub.docker.com/r/dmmcquay/starz/) contains the full README 
for the docker container. Please visit the link for complete details.

Here is an example of passing in all enviroment variables to the starz container:

```bash
$ docker run -d -p 80:8000 \
    -e STARZ_PORT=$STARZ_PORT \
    -e STARZ_HOSTNAME=$STARZ_HOSTNAME \
    -e STARZ_CLIENTID=$STARZ_CLIENTID \
    -e STARZ_CLIENTSECRET=$STARZ_CLIENTSECRET \
    -e STARZ_APITOKEN=$STARZ_APITOKEN \
    -e STARZ_COOKIESECRET=$STARZ_COOKIESECRET \
    dmmcquay/starz:latest
```

After you have this running, don't forget to update your github oauth call back
url to point to the ip, port, and route.

Deploying to GCE
==============================

In the gce folder you'll find my deployment scripts for deploying starz to GCE 
running a CoreOS instance.  It assumes that you have a correctly configured 
[gcloud tool enviroment](https://cloud.google.com/sdk/). The instance will, on 
boot, start a service running the docker container dmmcquay/starz:v1 and 
expose on port 80. 

Simply run:

```bash
$ bash deploy.sh
```

and the script with create a coreos instance with the running app.

After running deploy.sh, it will return a ip that you can go to 
your github oauth page and update the call back url with that ip plus the route:

```bash
http://$GCEIP/api/v0/github_oauth_cb/
```

The deploy.sh script will also make the external IP address you receive from 
GCE static, so you don't have to worry about it changing underneath you.

In the cloud-config.yaml file, you will have to add your github oauth clientid
and clientsecret.  You will also have to add your github api token if you want
to be able to have more than 60 requests per hour. Make sure you edit the 
cloud-config.yaml file, otherwise the values in there by default will cause
the application to fail to deploy.
