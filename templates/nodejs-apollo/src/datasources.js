import { RESTDataSource } from '@apollo/datasource-rest';

/**
 * Enhanced REST DataSource with caching and error handling
 * Generated for: {{PROJECT_NAME}}
 */

{{ range .BackendServices }}
export class {{ .Name | ToPascalCase }}API extends RESTDataSource {
  constructor() {
    super();
    this.baseURL = '{{ .BaseURL }}/';
  }


  // Forward authentication headers
  willSendRequest(_path, request) {
    const token = this.context.token;
    if (token) {
      request.headers['authorization'] = `Bearer ${token}`;
    }
  }

  async getAll() {
    return this.get('', {
      cacheOptions: { ttl: 60 }, // Cache for 60 seconds
    });
  }

  async getById(id) {
    return this.get(`${id}`, {
      cacheOptions: { ttl: 300 }, // Cache individual items longer
    });
  }

  async create(data) {
    return this.post('', { body: data });
  }

  async update(id, data) {
    return this.put(`${id}`, { body: data });
  }

  async delete(id) {
    return this.delete(`${id}`);
  }

  async healthCheck() {
    try {
      return await this.get('health');
    } catch (error) {
      return { status: 'unhealthy', error: error.message };
    }
  }
}
{{ end }}

// DataSources factory function
export const dataSources = () => ({
  {{ range .BackendServices }}
  {{ .Name | ToCamelCase }}API: new {{ .Name | ToPascalCase }}API(),
  {{ end }}
});
