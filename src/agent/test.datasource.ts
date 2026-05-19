import { registerAs } from '@nestjs/config';
import { DataSource, DataSourceOptions } from 'typeorm';

const config = {
  type: 'postgres',
  url: 'postgress://test:testpw@localhost:5431/testdb',
  entities: [],
  migrations: [],
};

export default registerAs('typeorm', () => config);
export const connectionSource = new DataSource(config as DataSourceOptions);
