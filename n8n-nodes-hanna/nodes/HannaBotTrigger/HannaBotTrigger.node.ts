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
    version: 2,
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
          {
            name: 'Server Events',
            value: 'server',
            description: 'Server-related events and numerics (001-099)',
          },
          {
            name: 'Channel Info Events',
            value: 'channel_info',
            description: 'Channel information events (321-366)',
          },
          {
            name: 'User Info Events',
            value: 'user_info',
            description: 'User information events (301-319)',
          },
          {
            name: 'Error Events',
            value: 'error',
            description: 'IRC error events and numerics (400-599)',
          },
          {
            name: 'MOTD Events',
            value: 'motd',
            description: 'Message of the Day events (372-376)',
          },
          {
            name: 'WHOIS Events',
            value: 'whois',
            description: 'WHOIS response events (311-319)',
          },
          {
            name: 'LIST Events',
            value: 'list',
            description: 'Channel list events (321-323)',
          },
          {
            name: 'Connection Events',
            value: 'connection',
            description: 'Connection status changes and registration',
          },
          {
            name: 'Capability Events',
            value: 'capability',
            description: 'IRC capability negotiation events',
          },
          {
            name: 'All Numeric Events',
            value: 'numeric',
            description: 'All IRC numeric events (001-999)',
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
      {
        displayName: 'Numeric Code Filter',
        name: 'numericFilter',
        type: 'string',
        default: '',
        description: 'Comma-separated list of specific IRC numeric codes to filter (e.g., 001,002,353). Leave empty for all numerics.',
        displayOptions: {
          show: {
            events: ['numeric'],
          },
        },
      },
      {
        displayName: 'Error Code Filter',
        name: 'errorCodeFilter',
        type: 'string',
        default: '',
        description: 'Comma-separated list of specific error codes to filter (e.g., 401,403,404). Leave empty for all errors.',
        displayOptions: {
          show: {
            events: ['error'],
          },
        },
      },
      {
        displayName: 'Minimum Severity',
        name: 'minSeverity',
        type: 'options',
        options: [
          {
            name: 'Info',
            value: 'info',
          },
          {
            name: 'Warning',
            value: 'warning',
          },
          {
            name: 'Error',
            value: 'error',
          },
          {
            name: 'Critical',
            value: 'critical',
          },
        ],
        default: 'info',
        description: 'Minimum severity level for events to trigger',
        displayOptions: {
          show: {
            events: ['error', 'server', 'connection'],
          },
        },
      },
    ],
  };

  async webhook(this: IWebhookFunctions): Promise<IWebhookResponseData> {
    const authToken = this.getNodeParameter('authToken') as string;
    const events = this.getNodeParameter('events') as string[];
    const channelFilter = this.getNodeParameter('channelFilter', '') as string;
    const userFilter = this.getNodeParameter('userFilter', '') as string;
    const includeBotMessages = this.getNodeParameter('includeBotMessages', false) as boolean;
    const numericFilter = this.getNodeParameter('numericFilter', '') as string;
    const errorCodeFilter = this.getNodeParameter('errorCodeFilter', '') as string;
    const minSeverity = this.getNodeParameter('minSeverity', 'info') as string;

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

    // Map event types to categories for enhanced filtering
    const eventType = eventData.eventType as string;
    const numericCode = eventData.numericCode as string;
    const errorCode = eventData.errorCode as string;
    const severity = eventData.severity as string || 'info';

    // Check if event matches selected types
    let eventMatches = false;
    
    for (const selectedEvent of events) {
      switch (selectedEvent) {
        case 'server':
          eventMatches = eventMatches || !!(numericCode && parseInt(numericCode) >= 1 && parseInt(numericCode) <= 99);
          break;
        case 'channel_info':
          eventMatches = eventMatches || !!(numericCode && parseInt(numericCode) >= 321 && parseInt(numericCode) <= 366);
          break;
        case 'user_info':
          eventMatches = eventMatches || !!(numericCode && parseInt(numericCode) >= 301 && parseInt(numericCode) <= 319);
          break;
        case 'error':
          eventMatches = eventMatches || !!(numericCode && parseInt(numericCode) >= 400 && parseInt(numericCode) <= 599);
          break;
        case 'motd':
          eventMatches = eventMatches || !!(numericCode && parseInt(numericCode) >= 372 && parseInt(numericCode) <= 376);
          break;
        case 'whois':
          eventMatches = eventMatches || !!(numericCode && parseInt(numericCode) >= 311 && parseInt(numericCode) <= 319);
          break;
        case 'list':
          eventMatches = eventMatches || !!(numericCode && parseInt(numericCode) >= 321 && parseInt(numericCode) <= 323);
          break;
        case 'connection':
          eventMatches = eventMatches || ['connect', 'disconnect', 'reconnect', 'registration'].includes(eventType);
          break;
        case 'capability':
          eventMatches = eventMatches || eventType.startsWith('cap_');
          break;
        case 'numeric':
          eventMatches = eventMatches || !!(numericCode && parseInt(numericCode) >= 1 && parseInt(numericCode) <= 999);
          break;
        default:
          eventMatches = eventMatches || eventType === selectedEvent;
      }
    }

    if (!eventMatches) {
      return {
        webhookResponse: {
          status: 200,
          body: { status: 'ignored', reason: 'Event type not monitored' },
        },
      };
    }

    // Filter by numeric codes if specified
    if (numericFilter && numericCode) {
      const allowedNumerics = numericFilter.split(',').map(n => n.trim());
      if (!allowedNumerics.includes(numericCode)) {
        return {
          webhookResponse: {
            status: 200,
            body: { status: 'ignored', reason: 'Numeric code not in filter' },
          },
        };
      }
    }

    // Filter by error codes if specified
    if (errorCodeFilter && errorCode) {
      const allowedErrorCodes = errorCodeFilter.split(',').map(e => e.trim());
      if (!allowedErrorCodes.includes(errorCode)) {
        return {
          webhookResponse: {
            status: 200,
            body: { status: 'ignored', reason: 'Error code not in filter' },
          },
        };
      }
    }

    // Filter by severity
    const severityLevels = ['info', 'warning', 'error', 'critical'];
    const minSeverityIndex = severityLevels.indexOf(minSeverity);
    const eventSeverityIndex = severityLevels.indexOf(severity);
    
    if (eventSeverityIndex !== -1 && eventSeverityIndex < minSeverityIndex) {
      return {
        webhookResponse: {
          status: 200,
          body: { status: 'ignored', reason: 'Event severity below minimum' },
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

    // Enhance event data with additional context
    const enhancedEventData = {
      ...eventData,
      webhookUrl: this.getNodeWebhookUrl('default'),
      receivedAt: new Date().toISOString(),
      // Add parsed information for easier workflow processing
      parsed: {
        isNumeric: !!numericCode,
        numericRange: numericCode ? Math.floor(parseInt(numericCode) / 100) * 100 : null,
        isError: errorCode !== undefined || (numericCode && parseInt(numericCode) >= 400),
        severity: severity,
        isChannelEvent: !!eventData.target && (eventData.target as string).startsWith('#'),
        isPrivateMessage: eventType === 'privmsg' && eventData.target && !(eventData.target as string).startsWith('#'),
      },
    };

    // Return the enhanced event data to the workflow
    return {
      webhookResponse: {
        status: 200,
        body: { status: 'processed' },
      },
      workflowData: [
        [
          {
            json: enhancedEventData,
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
