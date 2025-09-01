import {
  ITriggerFunctions,
  ITriggerResponse,
  INodeType,
  INodeTypeDescription,
  IWebhookResponseData,
  IWebhookFunctions,
  NodeConnectionType,
  IDataObject,
} from 'n8n-workflow';

export class HannaBotTrigger implements INodeType {
  description: INodeTypeDescription = {
    displayName: 'Hanna Bot Trigger',
    name: 'hannaBotTrigger',
    icon: 'fa:comments',
    group: ['trigger'],
    version: 1,
    subtitle: '={{$parameter["events"].join(", ") || "All IRC Events"}}',
    description: 'Triggers when IRC events occur in Hanna Bot',
    defaults: {
      name: 'Hanna Bot Trigger',
    },
    inputs: [],
    outputs: [NodeConnectionType.Main],
    webhooks: [
      {
        name: 'default',
        httpMethod: 'POST',
        responseMode: 'onReceived',
        path: 'webhook',
      },
    ],
    properties: [
      {
        displayName: 'Authentication Token',
        name: 'authToken',
        type: 'string',
        typeOptions: {
          password: true,
        },
        default: '',
        required: true,
        description: 'Token that the IRC bot must send to authenticate requests',
      },
      {
        displayName: 'IRC Events',
        name: 'events',
        type: 'multiOptions',
        options: [
          {
            name: 'Mention',
            value: 'mention',
            description: 'When the bot is mentioned in a message',
          },
          {
            name: 'Private Message',
            value: 'privmsg',
            description: 'All private messages (including channel messages)',
          },
          {
            name: 'Notice',
            value: 'notice',
            description: 'IRC notices received',
          },
          {
            name: 'Join',
            value: 'join',
            description: 'When someone joins a channel',
          },
          {
            name: 'Part',
            value: 'part',
            description: 'When someone leaves a channel',
          },
          {
            name: 'Quit',
            value: 'quit',
            description: 'When someone quits the IRC server',
          },
          {
            name: 'Kick',
            value: 'kick',
            description: 'When someone is kicked from a channel',
          },
          {
            name: 'Mode',
            value: 'mode',
            description: 'Channel or user mode changes',
          },
          {
            name: 'Nick',
            value: 'nick',
            description: 'When someone changes their nickname',
          },
          {
            name: 'Topic',
            value: 'topic',
            description: 'When channel topic is changed',
          },
        ],
        default: ['mention'],
        required: true,
        description: 'Select which IRC events should trigger this node',
      },
      {
        displayName: 'Filter Channels',
        name: 'channelFilter',
        type: 'string',
        default: '',
        description: 'Comma-separated list of channels to filter (e.g., #general,#dev). Leave empty for all channels.',
      },
      {
        displayName: 'Filter Users',
        name: 'userFilter',
        type: 'string',
        default: '',
        description: 'Comma-separated list of users to filter (e.g., alice,bob). Leave empty for all users.',
      },
      {
        displayName: 'Include Bot Messages',
        name: 'includeBotMessages',
        type: 'boolean',
        default: false,
        description: 'Whether to include messages from the bot itself',
      },
    ],
  };

  async webhook(this: IWebhookFunctions): Promise<IWebhookResponseData> {
    const authToken = this.getNodeParameter('authToken') as string;
    const events = this.getNodeParameter('events') as string[];
    const channelFilter = this.getNodeParameter('channelFilter', '') as string;
    const userFilter = this.getNodeParameter('userFilter', '') as string;
    const includeBotMessages = this.getNodeParameter('includeBotMessages', false) as boolean;

    const req = this.getRequestObject();
    const body = this.getBodyData();

    // Check authentication
    const providedToken = req.headers.authorization?.replace('Bearer ', '') || 
                         req.headers['x-auth-token'] || 
                         (body as IDataObject)?.authToken;

    if (!providedToken || providedToken !== authToken) {
      return {
        webhookResponse: {
          status: 401,
          body: { error: 'Invalid or missing authentication token' },
        },
      };
    }

    // Validate required fields
    const eventData = body as IDataObject;
    if (!eventData.eventType || !eventData.timestamp) {
      return {
        webhookResponse: {
          status: 400,
          body: { error: 'Missing required fields: eventType, timestamp' },
        },
      };
    }

    // Filter by event type
    if (!events.includes(eventData.eventType as string)) {
      return {
        webhookResponse: {
          status: 200,
          body: { status: 'ignored', reason: 'Event type not monitored' },
        },
      };
    }

    // Filter by channel
    if (channelFilter && eventData.target) {
      const allowedChannels = channelFilter.split(',').map(c => c.trim().toLowerCase());
      const targetChannel = (eventData.target as string).toLowerCase();
      if (!allowedChannels.includes(targetChannel)) {
        return {
          webhookResponse: {
            status: 200,
            body: { status: 'ignored', reason: 'Channel not in filter' },
          },
        };
      }
    }

    // Filter by user
    if (userFilter && eventData.sender) {
      const allowedUsers = userFilter.split(',').map(u => u.trim().toLowerCase());
      const senderUser = (eventData.sender as string).toLowerCase();
      if (!allowedUsers.includes(senderUser)) {
        return {
          webhookResponse: {
            status: 200,
            body: { status: 'ignored', reason: 'User not in filter' },
          },
        };
      }
    }

    // Filter bot messages
    if (!includeBotMessages && eventData.sender === eventData.botNick) {
      return {
        webhookResponse: {
          status: 200,
          body: { status: 'ignored', reason: 'Bot message filtered' },
        },
      };
    }

    // Return the event data to the workflow
    return {
      webhookResponse: {
        status: 200,
        body: { status: 'processed' },
      },
      workflowData: [
        [
          {
            json: {
              ...eventData,
              webhookUrl: this.getNodeWebhookUrl('default'),
              receivedAt: new Date().toISOString(),
            },
          },
        ],
      ],
    };
  }

  async trigger(this: ITriggerFunctions): Promise<ITriggerResponse> {
    return {
      closeFunction: async () => {
        // Cleanup function - nothing to clean up for webhook-based trigger
      },
    };
  }
}
