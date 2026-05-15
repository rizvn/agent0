This project demontrates building a simple agent with go lang

### Prerequisites:
- Go 1.26 or later

### To Build
Make is used to build the application
```bash
cd src 
make build
```

This will build the binary in the `bin` directory and copy env-example file

### To Run
- update `env-example` with settings to to your llm model
- rename `env-example` to `.env`
- run the binary

```bash
cd bin
./agent0
```

### Project Structure
- `main.go` is the entry point of the application. 
  - It loads environment variables
  - initializes the agent, and starts it.

- `config.go` defines the configuration struct and a function to load it from environment variables.

- `agent` - the agent directory contains the agent implementation. 
  - `agent.go` defines the agent struct and its methods. 

  - `tool` tools directory contains tools the agent can call
     - each tool implements the `Tool` interface defined in `tools.go`
     - `generic` directory contains a generic tool implementation 

     - Each tool has 3 parts
        - has a definition 
        - has an implementation
        - has a function to create the tool

      - Tools are registered with the agent in `agent.go`


### How it works
```mermaid
flowchart LR
    a[agent0] -- 1. Prompts user --> u[User]
    u[User] -- 2. What is the current time? --> a[agent0]
    a[agent0] -- 3. Calls llm with user prompt --> l[LLM]
    l -- 4. Returns response with tool call --> a
    a -- 5. Calls the tool --> a
    a -- 6. tells llm about tool result --> l
    l -- 7. Returns final response --> a
    a -- 8. Returns final response to user --> u
```