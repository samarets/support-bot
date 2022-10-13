# support-bot
Fast and simple **Telegram Support Bot** 🤖. It will help to connect the user and the support agent quickly and conveniently

## Features:
- All message types are forwarded
- Message replies work
- Adding/Removing support agents by a special command
- List of all support agents
- Post a support request in a separate chat
- Multilingualism and the ability to easily add your own language
- Error logging

### Admin Commands:
| Command        | Description                                |
|----------------|--------------------------------------------|
| `set_group`    | Set a group to receive notifications there |
| `add_support`  | Gives the user agent rights                |
| `del_support`  | Removes agent rights from the user         |
| `get_supports` | Gets the current list of agents            |

### Agent Commands:
| Command  | Description                         |
|----------|-------------------------------------|
| `break`  | Stop current conversation with user |

### User Commands:
| Command  | Description                 |
|----------|-----------------------------|
| `start`  | Get start message           |
| `get_id` | Corresponds to your user ID |

### How Bot Works:
1. Your client send question to your bot
2. The bot sends a notification request to your group or admin PM
3. Admin or Agents can confirm or decline appeal
4. If the appeal is declined, the user will receive a special message
5. If the request is confirmed, the user and the agent are connected to the joint chat

### Support Languages:
- Ukrainian
- English

> If you want to add another language, copy any relevant language, translate it and name the directory with your language and country code. For example: uk-UA

## Requirements:
- Docker `(or golang)`
- Docker-Compose `(or golang)`
- Makefile

## Installation:
1. Clone this repository `git clone https://github.com/samarets/support-bot.git`
2. Create .env file `cp  .env.example .env`
3. Paste your Bot Token and User ID to `.env`. Set default language
4. Run Bot:
    - With Docker: `make build`
    - Local with Go: 
      - `go mod download`
      - `go run cmd/bot/main.go`

## Images:
![Example Bot Image](https://raw.githubusercontent.com/samarets/support-bot/main/assets/user-agent.png)
