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
    subtitle: '={{$parameter["operation"] + ": " + $parameter["target"] || "IRC Bot Operations"}}',
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
