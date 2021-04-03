# matrix-bot

A plugin-based Matrix bot system written in Go.

# Current Plugins 🕹

- [github.js](https://github.com/Gurkengewuerz/matrix-bot/blob/main/plugins/github.js) responds to GitHub Event Webhooks
  like pipeline events or push events.
- [teamcity.js](https://github.com/Gurkengewuerz/matrix-bot/blob/main/plugins/teamcity.js) responds to [tcWebHooks](https://github.com/tcplugins/tcWebHooks) `Legacy Webhook (JSON)` requests
- [dice.js](https://github.com/Gurkengewuerz/matrix-bot/blob/main/plugins/dice.js) a simple dice rolling bot

# Configuration ✒

1. Create a new user `useradd -m matrix` and enter its home directory `cd /home/matrix`
2. Create a directory `mkdir matrix-bot` and enter it `cd matrix-bot`
3. Copy the build binaries and the `plugins/` directory into the folder
4. Copy `config.yaml.sample` to `config.yaml` and update to your liking
5. Make the binary executable `chmod +x matrix-bot`
6. Make sure every file has `matrix-bot` as the owner
--------
7. Set-Up a reverse proxy with Let's Encrypt
8. Install the `matrix-bot.service` systemd file to `/etc/systemd/system/`
9. Set-Up Pantalaimon for encryption support

# Commandline arguments 💻

- `-help` print help
- `-config` relative or absolute path of the `config.yaml`. Default: `${binary_dir}/config.yaml` 
- `-plugin` relative or absolute path of the `plugins/` folder. Default: `${binary_dir}/plugins/`

# How to build ⚙

1. Clone the repo:

```
git clone --recurse-submodules git@github.com:Gurkengewuerz/matrix-bot.git
cd matrix-bot
```

2. Build repository `go build`

⚠ Because we are using go-sqlite3 CGO is needed!

For cross-compiling from a newer Linux Distro (i.e. Windows WSL) to an older linux distro you can try `go build --ldflags '-linkmode external -extldflags "-static"'`

*Dev branch build artifacts for Linux/amd64 are available in the the GitHub Actions.*

# Encryption Support 🔐

Encryption is supported using the E2EE aware proxy daemon [pantalaimon](https://github.com/matrix-org/pantalaimon).  
Please follow the installation instructions in their repository. Afterwards you should set the Homeserver in the bot
config to pantalaimon. 

# Plugin Setup

### GitHub `github.js`
1. Active `github.js` in the `config.yaml`
2. (Re)start your bot
3. Test if plugin is loaded with `!gh ping`
4. Add your repository with `!gh add <username>/<repo>`
5. Get Webhook URL with `!gh webhook`
6. Add Webhook to your GitHub repository/project and select your wanted events

### TeamCity `teamcity.js`
1. Active `teamcity.js` in the `config.yaml`
2. (Re)start your bot
3. Test if plugin is loaded with `!tc ping`
4. Add your Build configuration (build config id) with `!tc add <build config id>`
5. Get Webhook URL with `!tc webhook`
6. Add a `Legacy Webhook (JSON)` to your project and select your wanted events
