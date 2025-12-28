import { ApolloServer } from '@apollo/server';
import { startStandaloneServer } from '@apollo/server/standalone';
import { typeDefs } from './schema.js';
import { resolvers } from './resolvers.js';
import { dataSources } from './datasources.js';

const server = new ApolloServer({
  typeDefs,
  resolvers,
});

const { url } = await startStandaloneServer(server, {
  context: async () => {
    const { cache } = server;
    return {
      dataSources: dataSources(cache),
    };
  },
  listen: { port: {{PORT}} },
});

console.log(`🚀  Apollo Server ready at: ${url}`);
