package main

import (
    "log"
    "os"
)

func init() {
    // Set test environment variables for the enhanced bot
    os.Setenv("API_PORT", "8084")
    os.Setenv("API_TOKEN", "test-token-12345")
    os.Setenv("IRC_ADDR", "irc.libera.chat:6697")
    os.Setenv("IRC_NICK", "HannaTestBot")
    os.Setenv("IRC_USER", "hannatest")
    os.Setenv("IRC_NAME", "Hanna Test Bot Enhanced")
    
    log.Println("Environment variables set for enhanced IRC bot testing")
}


