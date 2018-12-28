## Mi AI
Mi AI is an utility messenger bot which will help you grab some stuff from the web.

## Deployment on Debian based distros
1. Clone the repository on your server
2. You need SSL certificate. You can get one for free using [Let's encrypt](https://letsencrypt.org)
3. Once you got your certificate, create a `.env` file inside the project directory. This file should contain the following informations
    ````
    ENVIRONMENT=local
    CERT_PATH=<path_to_your_certificate>/fullchain.pem
    CERT_KEY_PATH=<path_to_your_certificate_key>/privkey.pem
    VERIFY_TOKEN=<your_verify_token>
    ````
4. Next step is to run the web service. Inside the project directory, run the build command (this will give you an binary file named `server`)
    ````
    $ go build server.go
    ````
5. I recommend using [supervisor](http://supervisord.org) to run the server as system service. You can of course use systemd if you want.

    * Install supervisor with
        ```
        $ sudo apt-get install supervisor
        ```

    * Create the following configuration file `/etc/supservisor/conf.d/mi-ai.conf` with the following content
    
        ```
        [program:mi_ai]
        directory=/home/<user-name>/go/src/mi-ai
        command=/home/<user-name>/go/src/mi-ai/server
        autostart=true
        autorestart=true
        stderr_logfile=/var/log/mi-ai.err.log
        stdout_logfile=/var/log/mi-ai.out.log
        ```
    
    * Run the following commands
        ```
        $ supervisorctl reread
        $ supervisorctl update
        ```
    
    For more information about supervisor configuration, go [there](https://www.digitalocean.com/community/tutorials/how-to-install-and-manage-supervisor-on-ubuntu-and-debian-vps)


