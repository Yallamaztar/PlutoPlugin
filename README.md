# PlutoPlugin (WIP)

![plutoplugin](.github/banner.png)

a modular server management and plugin framework for **Call of Duty: Black Ops II [(Plutonium T6)](https://plutonium.pw/)**
It aims to provide a flexible foundation for economy systems, administrative tools, and third-party **integrations, including Discord and [IW4M-Admin](https://github.com/RaidMax/IW4M-Admin)**

| This project is work in progress

## Features
- **Economy**: 
    - Player wallets & banking
    - Gambling systems
    - Ingame Shop

- **Administration**
    - Player management
    - Server controls
    - Map rotation
    - Custom command & level system

- **Discord Integration (optional)**
    - Economy commands
    - Administrative actions
    - Event notifications 

- **IW4M-Admin Integration (optional)**
    - Role & permission syncing
    - Ban management
    - Administrative actions
    - Session-based & performance-based rewards
    
    **For help with IW4M-Admin integration [click here](https://github.com/Yallamaztar/iw4m?tab=readme-ov-file#setting-up-iw4m)**

---

### Installation & Building
1. **Clone the repository**
    ```shell
    git clone https://github.com/Yallamaztar/PlutoPlugin.git
    cd PlutoPlugin
    ```

2. **Build the plugin**
    ```shell
    go build -ldflags="-s -w -buildid=" -trimpath cmd\plugin\main.go
    ```

3. **Place the GSC scripts in the your scripts/ dir**

4. **Run the plugin**
    ```shell
    main.exe
    ```

---

### TODO:
- Finish all command implementations
- Finish discord bot
- Improve event handling
- Documentation & examples
- Add custom bots (maybe)