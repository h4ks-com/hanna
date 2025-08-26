import {
  IAuthenticateGeneric,
  ICredentialTestRequest,
  ICredentialType,
  INodeProperties,
} from 'n8n-workflow';

export class HannaBotApi implements ICredentialType {
  name = 'hannaBotApi';
  displayName = 'Hanna Bot API';
  description = 'Authenticate with Hanna IRC Bot REST API';
  documentationUrl = 'https://github.com/h4ks-com/hanna';
  properties: INodeProperties[] = [
    {
      displayName: 'API URL',
      name: 'apiUrl',
      type: 'string',
      default: 'http://localhost:8080',
      description: 'Base URL of your Hanna Bot API instance (e.g., https://bot.example.com)',
      required: true,
    },
    {
      displayName: 'API Token',
      name: 'apiToken',
      type: 'string',
      typeOptions: {
        password: true,
      },
      default: '',
      description: 'Bearer token for API authentication. Set this as API_TOKEN when running Hanna Bot.',
      required: true,
    },
  ];

  authenticate: IAuthenticateGeneric = {
    type: 'generic',
    properties: {
      headers: {
        Authorization: '=Bearer {{$credentials.apiToken}}',
      },
    },
  };

  test: ICredentialTestRequest = {
    request: {
      baseURL: '={{$credentials.apiUrl}}',
      url: '/api/state',
      method: 'GET',
    },
  };
}