# n8n Hanna Bot Node - LIST and WHOIS Operations

## Overview

The Hanna Bot n8n node now supports IRC channel listing and user information gathering through the new `List Channels` and `Get User Info (WHOIS)` operations.

## New Operations

### List Channels

**Operation**: `List Channels`
**Method**: GET
**Description**: Retrieves a complete list of all IRC channels on the network.

**Response Structure**:
```json
{
  "success": true,
  "operation": "list",
  "channelCount": 42,
  "totalUsers": 1337,
  "channels": [
    {
      "channel": "#general",
      "users": "156",
      "topic": "General discussion channel"
    },
    {
      "channel": "#bots", 
      "users": "23",
      "topic": "Bot testing area"
    }
  ],
  "largestChannels": [
    {
      "channel": "#general",
      "users": "156", 
      "topic": "General discussion channel"
    }
  ],
  "response": {
    "channels": [...],
    "count": 42
  },
  "timestamp": "2025-09-04T10:00:00.000Z"
}
```

**Key Fields**:
- `channelCount`: Total number of channels
- `totalUsers`: Sum of all users across all channels  
- `channels`: Array of channel objects with name, user count, and topic
- `largestChannels`: Top 5 channels by user count

### Get User Info (WHOIS)

**Operation**: `Get User Info (WHOIS)`
**Method**: POST
**Parameters**: 
- `Nickname`: IRC nickname to query

**Response Structure**:
```json
{
  "success": true,
  "operation": "whois",
  "queriedNick": "username",
  "userInfo": {
    "nick": "username",
    "user": "user", 
    "host": "example.com",
    "realName": "Real Name",
    "server": "irc.libera.chat",
    "serverInfo": "Stockholm, Sweden",
    "isOperator": false,
    "idleSeconds": 42,
    "channels": "@#ops +#general #random"
  },
  "userChannels": [
    {
      "channel": "#ops",
      "modes": "@",
      "isOperator": true,
      "hasVoice": false,
      "isHalfOp": false
    },
    {
      "channel": "#general", 
      "modes": "+",
      "isOperator": false,
      "hasVoice": true,
      "isHalfOp": false
    },
    {
      "channel": "#random",
      "modes": "",
      "isOperator": false,
      "hasVoice": false,
      "isHalfOp": false
    }
  ],
  "response": {
    "nick": "username",
    "user": "user",
    "host": "example.com",
    "real_name": "Real Name",
    "server": "irc.libera.chat",
    "server_info": "Stockholm, Sweden", 
    "operator": false,
    "idle_seconds": "42",
    "channels": "@#ops +#general #random",
    "raw_data": [...]
  },
  "timestamp": "2025-09-04T10:00:00.000Z"
}
```

**Key Fields**:
- `queriedNick`: The nickname that was queried
- `userInfo`: Structured user information with boolean flags
- `userChannels`: Parsed array of channels with mode information
- `userChannels[].isOperator`: True if user has @ (operator) in this channel
- `userChannels[].hasVoice`: True if user has + (voice) in this channel
- `userChannels[].isHalfOp`: True if user has % (half-operator) in this channel

## Example Workflows

### 1. Channel Discovery Workflow

**Use Case**: Find active channels and monitor popular ones.

```json
{
  "nodes": [
    {
      "name": "List Channels",
      "type": "n8n-nodes-hanna.hannaBot", 
      "parameters": {
        "operation": "list"
      }
    },
    {
      "name": "Filter Large Channels",
      "type": "n8n-nodes-base.filter",
      "parameters": {
        "conditions": {
          "number": [
            {
              "value1": "={{$json.channels}}",
              "operation": "length",
              "value2": 10
            }
          ]
        }
      }
    },
    {
      "name": "Process Each Channel", 
      "type": "n8n-nodes-base.splitInBatches",
      "parameters": {
        "batchSize": 1,
        "options": {
          "destinationKey": "channel"
        }
      }
    },
    {
      "name": "Join Popular Channels",
      "type": "n8n-nodes-hanna.hannaBot",
      "parameters": {
        "operation": "join",
        "target": "={{$json.channel.channel}}"
      }
    }
  ]
}
```

### 2. User Authentication Workflow

**Use Case**: Verify new users and welcome them based on their information.

```json
{
  "nodes": [
    {
      "name": "User Joined Trigger",
      "type": "n8n-nodes-hanna.hannaBotTrigger",
      "parameters": {
        "events": ["join"],
        "channels": ["#general"]
      }
    },
    {
      "name": "Get User Info",
      "type": "n8n-nodes-hanna.hannaBot",
      "parameters": {
        "operation": "whois", 
        "whoisNick": "={{$json.sender}}"
      }
    },
    {
      "name": "Check if Operator",
      "type": "n8n-nodes-base.if",
      "parameters": {
        "conditions": {
          "boolean": [
            {
              "value1": "={{$json.userInfo.isOperator}}",
              "value2": true
            }
          ]
        }
      }
    },
    {
      "name": "Welcome Operator",
      "type": "n8n-nodes-hanna.hannaBot",
      "parameters": {
        "operation": "send",
        "target": "={{$('User Joined Trigger').item.json.target}}",
        "message": "Welcome {{$('User Joined Trigger').item.json.sender}}! Thanks for joining, operator! 👑"
      }
    },
    {
      "name": "Welcome Regular User", 
      "type": "n8n-nodes-hanna.hannaBot",
      "parameters": {
        "operation": "send",
        "target": "={{$('User Joined Trigger').item.json.target}}",
        "message": "Welcome {{$('User Joined Trigger').item.json.sender}} from {{$json.userInfo.host}}! 👋"
      }
    }
  ]
}
```

### 3. Channel Statistics Workflow

**Use Case**: Generate regular reports about IRC network activity.

```json
{
  "nodes": [
    {
      "name": "Schedule Daily",
      "type": "n8n-nodes-base.cron",
      "parameters": {
        "triggerTimes": {
          "hour": 9,
          "minute": 0
        }
      }
    },
    {
      "name": "Get Channel List",
      "type": "n8n-nodes-hanna.hannaBot",
      "parameters": {
        "operation": "list"
      }
    },
    {
      "name": "Generate Report",
      "type": "n8n-nodes-base.code",
      "parameters": {
        "jsCode": "const data = items[0].json;\nconst channels = data.channels;\nconst total = data.totalUsers;\nconst count = data.channelCount;\n\nconst report = `📊 Daily IRC Network Report\\n\\n` +\n  `📈 Total Channels: ${count}\\n` +\n  `👥 Total Users: ${total}\\n` +\n  `📊 Average Users per Channel: ${Math.round(total / count)}\\n\\n` +\n  `🔥 Top 5 Channels:\\n` +\n  data.largestChannels.map((ch, i) => \n    `${i + 1}. ${ch.channel} (${ch.users} users)\\n`\n  ).join('') +\n  `\\n📝 Topics with keywords:\\n` +\n  channels.filter(ch => \n    ch.topic && (ch.topic.includes('bot') || ch.topic.includes('dev'))\n  ).map(ch => \n    `• ${ch.channel}: ${ch.topic}\\n`\n  ).join('');\n\nreturn [{ json: { report } }];"
      }
    },
    {
      "name": "Send Report",
      "type": "n8n-nodes-hanna.hannaBot",
      "parameters": {
        "operation": "send",
        "target": "#reports",
        "message": "={{$json.report}}"
      }
    }
  ]
}
```

### 4. User Channel Analysis

**Use Case**: Analyze user's channel membership for moderation.

```json
{
  "nodes": [
    {
      "name": "Manual Trigger",
      "type": "n8n-nodes-base.manualTrigger"
    },
    {
      "name": "Get User Info",
      "type": "n8n-nodes-hanna.hannaBot",
      "parameters": {
        "operation": "whois",
        "whoisNick": "suspicious_user"
      }
    },
    {
      "name": "Analyze Channels",
      "type": "n8n-nodes-base.code", 
      "parameters": {
        "jsCode": "const userChannels = items[0].json.userChannels || [];\nconst operatorChannels = userChannels.filter(ch => ch.isOperator);\nconst voiceChannels = userChannels.filter(ch => ch.hasVoice);\nconst regularChannels = userChannels.filter(ch => !ch.isOperator && !ch.hasVoice);\n\nconst analysis = {\n  totalChannels: userChannels.length,\n  operatorIn: operatorChannels.length,\n  voiceIn: voiceChannels.length,\n  regularIn: regularChannels.length,\n  operatorChannels: operatorChannels.map(ch => ch.channel),\n  suspiciousPattern: userChannels.length > 20 || operatorChannels.length > 5,\n  trustLevel: operatorChannels.length > 0 ? 'high' : \n             voiceChannels.length > 2 ? 'medium' : 'low'\n};\n\nreturn [{ json: { ...items[0].json, analysis } }];"
      }
    },
    {
      "name": "Send Analysis",
      "type": "n8n-nodes-hanna.hannaBot",
      "parameters": {
        "operation": "send", 
        "target": "#moderation",
        "message": "User Analysis: {{$json.queriedNick}}\\nChannels: {{$json.analysis.totalChannels}}\\nOperator in: {{$json.analysis.operatorIn}}\\nTrust Level: {{$json.analysis.trustLevel}}\\nSuspicious: {{$json.analysis.suspiciousPattern}}"
      }
    }
  ]
}
```

## Advanced Usage Tips

### 1. Error Handling

Both operations can fail if the bot is not connected or if the IRC server doesn't respond:

```json
{
  "name": "Handle Errors",
  "type": "n8n-nodes-base.if",
  "parameters": {
    "conditions": {
      "boolean": [
        {
          "value1": "={{$json.success}}",
          "value2": false
        }
      ]
    }
  }
}
```

### 2. Rate Limiting

Be mindful of IRC server rate limits:

```json
{
  "name": "Delay Between Requests",
  "type": "n8n-nodes-base.wait",
  "parameters": {
    "amount": 2,
    "unit": "seconds"
  }
}
```

### 3. Data Processing

Use the structured data for easy filtering:

```javascript
// Filter channels by user count
const largeChannels = $json.channels.filter(ch => parseInt(ch.users) > 50);

// Check if user is operator anywhere
const isGlobalOp = $json.userChannels.some(ch => ch.isOperator);

// Get user's server location
const serverLocation = $json.userInfo.serverInfo;
```

## Integration with Existing Workflows

The new operations work seamlessly with existing Hanna Bot functionality:

1. **List channels** → **Join interesting ones** → **Monitor for activity**
2. **User joins** → **WHOIS user** → **Apply appropriate permissions**
3. **Regular channel audit** → **Generate reports** → **Notify administrators**

## Performance Considerations

- **LIST**: Can be resource-intensive on large networks, consider caching results
- **WHOIS**: Rate-limited by most IRC servers, use sparingly
- **Timeout**: Both operations have 10-second timeout, suitable for most networks
- **Caching**: Consider storing results temporarily to avoid repeated requests
