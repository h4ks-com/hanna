import {
  IExecuteFunctions,
  INodeExecutionData,
  INodeType,
  INodeTypeDescription,
  NodeConnectionType,
  NodeOperationError,
} from 'n8n-workflow';

export class HannaBot implements INodeType {
  description: INodeTypeDescription = {
    displayName: 'Hanna Bot',
    name: 'hannaBot',
    icon: 'fa:comments',
    group: ['transform'],
    version: 1,
    subtitle: '={{$parameter["operation"] === "whois" ? "WHOIS: " + $parameter["whoisNick"] : $parameter["operation"] === "list" ? "List Channels" : $parameter["operation"] + ": " + $parameter["target"] || "IRC Bot Operations"}}',
    description: 'Interact with Hanna IRC Bot via its REST API',
    usableAsTool: true,
    defaults: {
      name: 'Hanna Bot',
    },
    inputs: [NodeConnectionType.Main],
    outputs: [NodeConnectionType.Main],
    credentials: [
      {
        name: 'hannaBotApi',
        required: true,
      },
    ],
    properties: [
      {
        displayName: 'Operation',
        name: 'operation',
        type: 'options',
        noDataExpression: true,
        options: [
          {
            name: 'Send Message',
            value: 'send',
            description: 'Send a message to a channel or user',
            action: 'Send a message',
          },
          {
            name: 'Send Notice',
            value: 'notice',
            description: 'Send a notice to a channel or user',
            action: 'Send a notice',
          },
          {
            name: 'Join Channel',
            value: 'join',
            description: 'Join an IRC channel',
            action: 'Join a channel',
          },
          {
            name: 'Part Channel',
            value: 'part',
            description: 'Leave an IRC channel',
            action: 'Leave a channel',
          },
          {
            name: 'Get Status',
            value: 'status',
            description: 'Get bot connection status and channels',
            action: 'Get bot status',
          },
          {
            name: 'Change Nick',
            value: 'nick',
            description: 'Change bot nickname',
            action: 'Change nickname',
          },
          {
            name: 'Send Raw Command',
            value: 'raw',
            description: 'Send raw IRC command',
            action: 'Send raw IRC command',
          },
          {
            name: 'List Channels',
            value: 'list',
            description: 'Get list of all IRC channels on the network',
            action: 'List IRC channels',
          },
          {
            name: 'Get User Info (WHOIS)',
            value: 'whois',
            description: 'Get detailed information about a specific user',
            action: 'Get user information',
          },
        ],
        default: 'send',
        required: true,
      },
      // Target field for send, notice, join, part operations
      {
        displayName: 'Target',
        name: 'target',
        type: 'string',
        default: '',
        placeholder: '#general',
        required: true,
        description: 'IRC channel (e.g., #general) or username to target',
        displayOptions: {
          show: {
            operation: ['send', 'notice', 'join', 'part'],
          },
        },
      },
      // Nick field for whois operation
      {
        displayName: 'Nickname',
        name: 'whoisNick',
        type: 'string',
        default: '',
        placeholder: 'username',
        required: true,
        description: 'IRC nickname to get information about',
        displayOptions: {
          show: {
            operation: ['whois'],
          },
        },
      },
      // Message field for send and notice operations
      {
        displayName: 'Message',
        name: 'message',
        type: 'string',
        typeOptions: {
          rows: 3,
        },
        default: '',
        required: true,
        description: 'The message content to send. Can be AI-generated content.',
        displayOptions: {
          show: {
            operation: ['send', 'notice'],
          },
        },
      },
      // Part reason (optional)
      {
        displayName: 'Reason',
        name: 'reason',
        type: 'string',
        default: '',
        description: 'Optional reason for leaving the channel',
        displayOptions: {
          show: {
            operation: ['part'],
          },
        },
      },
      // Nick field for nick operation
      {
        displayName: 'New Nickname',
        name: 'nick',
        type: 'string',
        default: '',
        required: true,
        description: 'New nickname for the bot',
        displayOptions: {
          show: {
            operation: ['nick'],
          },
        },
      },
      // Raw command field
      {
        displayName: 'IRC Command',
        name: 'command',
        type: 'string',
        default: '',
        required: true,
        description: 'Raw IRC command to send (e.g., "PRIVMSG #channel :Hello world")',
        displayOptions: {
          show: {
            operation: ['raw'],
          },
        },
      },
    ],
  };

  async execute(this: IExecuteFunctions): Promise<INodeExecutionData[][]> {
    const items = this.getInputData();
    const returnData: INodeExecutionData[] = [];

    for (let i = 0; i < items.length; i++) {
      try {
        const operation = this.getNodeParameter('operation', i) as string;
        const credentials = await this.getCredentials('hannaBotApi', i);

        const baseUrl = credentials.apiUrl as string;
        const token = credentials.apiToken as string;

        const headers = {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        };

        let endpoint = '';
        let method = 'GET';
        let body: any = undefined;

        switch (operation) {
          case 'send':
            endpoint = '/api/send';
            method = 'POST';
            body = {
              target: this.getNodeParameter('target', i) as string,
              message: this.getNodeParameter('message', i) as string,
            };
            break;

          case 'notice':
            endpoint = '/api/notice';
            method = 'POST';
            body = {
              target: this.getNodeParameter('target', i) as string,
              message: this.getNodeParameter('message', i) as string,
            };
            break;

          case 'join':
            endpoint = '/api/join';
            method = 'POST';
            body = {
              channel: this.getNodeParameter('target', i) as string,
            };
            break;

          case 'part':
            endpoint = '/api/part';
            method = 'POST';
            body = {
              channel: this.getNodeParameter('target', i) as string,
              reason: this.getNodeParameter('reason', i, '') as string,
            };
            break;

          case 'status':
            endpoint = '/api/state';
            method = 'GET';
            break;

          case 'nick':
            endpoint = '/api/nick';
            method = 'POST';
            body = {
              nick: this.getNodeParameter('nick', i) as string,
            };
            break;

          case 'raw':
            endpoint = '/api/raw';
            method = 'POST';
            body = {
              line: this.getNodeParameter('command', i) as string,
            };
            break;

          case 'list':
            endpoint = '/api/list';
            method = 'GET';
            break;

          case 'whois':
            endpoint = '/api/whois';
            method = 'POST';
            body = {
              nick: this.getNodeParameter('whoisNick', i) as string,
            };
            break;

          default:
            throw new NodeOperationError(this.getNode(), `Invalid operation: ${operation}`, {
              itemIndex: i,
            });
        }

        const requestOptions = {
          method: method as 'GET' | 'POST',
          body: body ? JSON.stringify(body) : undefined,
          headers,
        };

        const response = await this.helpers.request(`${baseUrl}${endpoint}`, requestOptions);

        const responseData: any = {
          success: true,
          operation,
          response: typeof response === 'string' ? JSON.parse(response) : response,
          timestamp: new Date().toISOString(),
        };

        // Add operation-specific data to response
        if (operation === 'send' || operation === 'notice') {
          responseData.target = this.getNodeParameter('target', i);
          responseData.message = this.getNodeParameter('message', i);
        } else if (operation === 'join' || operation === 'part') {
          responseData.channel = this.getNodeParameter('target', i);
          if (operation === 'part') {
            responseData.reason = this.getNodeParameter('reason', i, '');
          }
        } else if (operation === 'nick') {
          responseData.newNick = this.getNodeParameter('nick', i);
        } else if (operation === 'raw') {
          responseData.command = this.getNodeParameter('command', i);
        } else if (operation === 'list') {
          // For list operation, include channel count and make channels easily accessible
          const channels = responseData.response.channels || [];
          responseData.channelCount = responseData.response.count || channels.length;
          responseData.channels = channels;
          
          // Add summary statistics
          const totalUsers = channels.reduce((sum: number, channel: any) => {
            return sum + parseInt(channel.users || '0', 10);
          }, 0);
          responseData.totalUsers = totalUsers;
          
          // Find largest channels
          const sortedChannels = [...channels].sort((a: any, b: any) => {
            return parseInt(b.users || '0', 10) - parseInt(a.users || '0', 10);
          });
          responseData.largestChannels = sortedChannels.slice(0, 5);
        } else if (operation === 'whois') {
          // For whois operation, add the queried nick and structured data
          responseData.queriedNick = this.getNodeParameter('whoisNick', i);
          
          // Extract key information for easy access
          const whoisData = responseData.response;
          responseData.userInfo = {
            nick: whoisData.nick,
            user: whoisData.user,
            host: whoisData.host,
            realName: whoisData.real_name,
            server: whoisData.server,
            serverInfo: whoisData.server_info,
            isOperator: whoisData.operator || false,
            idleSeconds: whoisData.idle_seconds ? parseInt(whoisData.idle_seconds, 10) : null,
            channels: whoisData.channels,
          };
          
          // Parse channels into array for easier processing
          if (whoisData.channels) {
            const channelList = whoisData.channels.split(' ').filter((ch: string) => ch.trim());
            responseData.userChannels = channelList.map((ch: string) => {
              const modes = [];
              let channel = ch;
              
              // Extract channel modes (@, +, %, etc.)
              while (channel.length > 0 && ['@', '+', '%', '&', '~'].includes(channel[0])) {
                modes.push(channel[0]);
                channel = channel.substring(1);
              }
              
              return {
                channel,
                modes: modes.join(''),
                isOperator: modes.includes('@'),
                hasVoice: modes.includes('+'),
                isHalfOp: modes.includes('%'),
              };
            });
          }
        }

        returnData.push({
          json: responseData,
          pairedItem: { item: i },
        });

      } catch (error) {
        if (this.continueOnFail()) {
          returnData.push({
            json: {
              success: false,
              error: error instanceof Error ? error.message : String(error),
              operation: this.getNodeParameter('operation', i, 'unknown'),
              timestamp: new Date().toISOString(),
            },
            pairedItem: { item: i },
          });
          continue;
        }

        throw new NodeOperationError(this.getNode(), error instanceof Error ? error : String(error), {
          itemIndex: i,
        });
      }
    }

    return [returnData];
  }
}
