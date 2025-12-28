/**
 * GraphQL Resolvers
 * Project: {{PROJECT_NAME}}
 */

export const resolvers = {
  Query: {
    health: async (_, __, { dataSources }) => {
      // Aggregate health checks from all services
      const serviceHealthChecks = await Promise.allSettled([
        {{ range .BackendServices }}
        dataSources.{{ .Name | ToCamelCase }}API.healthCheck(),
        {{ end }}
      ]);

      const services = [
        {{ range $index, $ := .BackendServices }}
        {
          name: '{{ .Name }}',
          status: serviceHealthChecks[{{ $index }}].status === 'fulfilled' ? 
            serviceHealthChecks[{{ $index }}].value.status : 'unhealthy',
          url: '{{ .BaseURL }}'
        },
        {{ end }}
      ];

      const overallStatus = services.every(s => s.status === 'healthy') ? 'healthy' : 'degraded';

      return {
        status: overallStatus,
        services
      };
    },

    {{ range .BackendServices }}
    {{ .Name | ToCamelCase }}: async (_, { id }, { dataSources }) => {
      return await dataSources.{{ .Name | ToCamelCase }}API.getById(id);
    },

    {{ .Name | ToCamelCase }}List: async (_, __, { dataSources }) => {
      return await dataSources.{{ .Name | ToCamelCase }}API.getAll();
    },
    {{ end }}
  },

  Mutation: {
    {{ range .BackendServices }}
    create{{ .Name | ToPascalCase }}: async (_, { input }, { dataSources }) => {
      return await dataSources.{{ .Name | ToCamelCase }}API.create(input);
    },

    update{{ .Name | ToPascalCase }}: async (_, { id, input }, { dataSources }) => {
      return await dataSources.{{ .Name | ToCamelCase }}API.update(id, input);
    },

    delete{{ .Name | ToPascalCase }}: async (_, { id }, { dataSources }) => {
      await dataSources.{{ .Name | ToCamelCase }}API.delete(id);
      return true;
    },
    {{ end }}
  },
};
