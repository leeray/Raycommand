Ray Command: AI-Powered Terminal Assistant

![ray-command-demo](./docs/imgs/1.png)

🚀 Overview
Ray Command is an intelligent CLI tool that leverages AI to help you generate and execute terminal commands quickly and efficiently. By using natural language prompts, you can get precise bash commands generated by an AI model.

✨ Features
 - Generate bash commands using AI
 - Direct command execution
 - Support for multiple AI providers ([DeepSeek](http://deepseek.com/) currently implemented)


🛠 Prerequisites
 - Go 1.22.10 or higher
 - [DeepSeek](http://deepseek.com/) & [OpenAI](https://openai.com/) API Key
 - bash shell


🔧 Installation
1. Clone the repository:
```bash
git clone https://github.com/yourusername/raycommand.git
cd raycommand
```
2. Install dependencies:
```bash
go mod tidy
```
3. Set up environment variables: Create a .env file in the project root and add your [DeepSeek](http://deepseek.com/)  API key:
```bash
DEEPSEEK_API_KEY=your_api_key_here
```

💡 Usage
Run the command with your natural language request:
```bash
raycommand "Generate a bash command to list all files in the current directory"
```


⚖️ License
Distributed under the MIT License. See LICENSE for more information.
