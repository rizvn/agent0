This project demontrates building a simple ai agent with go lang

### Prerequisites
- Go 1.26 or later


### To Build 
Create bin folder at same level as src
```bash
mkdir bin
```

Folder structure should look like this
```
|-- bin
|-- src
    |-- main.go
```

Build commands
```
cd src
go build -o ../bin/agent0 main.go
```


Copy env-example to bin
```bash
cp env-example ../bin/.env
```

Update .env with your llm settings for example  
```
LLM_BASE_URL=your-llm-base-url
LLM_MODEL=your-llm-model
LLM_API_KEY=your-llm-api-key
```

### To Run
```bash
cd bin
./agent0
``` 


#### To Run in Non interactive mode
You can also run the agent in non interactive mode by passing a message as an argument
```bash
cd bin
./agent0 -p "What is the current time?"  
```



#### To Run in Non interactive mode
You can also run the agent in non interactive mode by passing a message as an argument
```bash
cd bin
./agent0 -p "What is the current time?"  
```


### How it works
```mermaid
sequenceDiagram
autonumber

actor user
participant agent
participant llm
user ->> agent : User enters a message<br/> e.g. "What is the current time?"
agent ->> llm : Call LLM with user message <br/> e.g. "What is the current time?" <br> And list of tool definitions
llm ->> agent : Responds with tool call <br/> e.g. "bash date"
agent ->> agent : Call tool <br/> e.g. call bashtool and executes "date" command
agent ->> llm : Send tool result <br/> e.g. "The result of the bash date command is: Sat Sep 30 12:34:56 UTC 2023"
llm ->> agent : Returns final response <br/> e.g. "The current time is: Sat Sep 30 12:34:56 UTC 2023" 
agent ->> user : Returns final response to user <br/> e.g. "The current time is: Sat Sep 30 12:34:56 UTC 2023"
```

### Project Structure
- `main.go` is the entry point of the application. 
  - It loads environment variables
  - initializes the agent, and starts it.


- `agent` - the agent directory contains the agent implementation. 
  - `config.go` defines the configuration struct and a function to load it from environment variables.

  - `agent.go` defines the agent struct and its methods. 

  - `tool` tools directory contains tools the agent can call
     - each tool implements the `Tool` interface defined in `tools.go`
     - `generic` directory contains a collection of generic tools 

     - Each tool has 3 parts:
        - a function to create an instance of the tool e.g `New<ToolName>`
        - a method to provide tool definition e.g `Defintion()`
        - a method to call the tool e.g `Call(...)`

      - Tools are registered with the agent in `agent.go`

