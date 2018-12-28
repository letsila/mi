## Mimi AI
Mimi AI is an utility messenger bot which will help you doing some stuff on the web.

## Deployment on Debian based distros
1. Clone the repository on your server
2. You need SSL certificate. You can get one for free using [Let's encrypt](https://letsencrypt.org)
3. Once you are set up with your domain and your SSL certificate, next step is to run the web service. Inside the project directory, run the build command (this will give you an binary file named `server`)
```bash
$ go build server.go
```
4. I recommend using [supervisor](http://supervisord.org) to run the server as system service. You can of courser use systemd if you want.

    3.1 Install supervisor with
    ```
    $ sudo apt-get install supervisor
    ```

    3.2 Create the following configuration file `/etc/supservisor/conf.d/mimi-ai.conf` with the following content

    > [program:mimi_ai]
    > directory=/home/<user-name>/go/src/mimi-ai
    > command=/home/<user-name>/go/src/mimi-ai/server
    > autostart=true
    > autorestart=true
    > stderr_logfile=/var/log/mimi-ai.err.log
    > stdout_logfile=/var/log/mimi-ai.out.log
    
    3.3 Run the following commands
    ```
    $ supervisorctl reread
    $ supervisorctl update
    ```
    
    For more information about supervisor configuration, go [there](https://www.digitalocean.com/community/tutorials/how-to-install-and-manage-supervisor-on-ubuntu-and-debian-vps)


