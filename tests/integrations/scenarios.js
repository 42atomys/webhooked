import { randomString } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';
import { describe, expect } from 'https://jslib.k6.io/k6chaijs/4.3.4.3/index.js';
import { redisClient, session } from './options.js';

const randomName = randomString(10);

export const scenarios = [
  {
    name: 'basic-usage',
    description: 'should return 200 with the payload not formatted.',
    payload: { 
      message: `Hello basic, ${randomName}!`,
    },
    expected: {
      message: `Hello basic, ${randomName}!`,
    },
    expectedResponse: ''
  },
  {
    name: 'basic-formatted-usage',
    description: 'should return 200 with a basic formatting.',
    payload: {
      message: `Hello formatted, ${randomName}!`,
    },
    expected: {
      "contentType": "application/json",
      data: {
        message: `Hello formatted, ${randomName}!`,
      }
    },
    expectedResponse: ''
  },
  {
    name: 'basic-response',
    description: 'should return 200 with a response asked.',
    payload: {
      id: randomName,
    },
    expected: {
      id: randomName,
    },
    expectedResponse: randomName
  },
  {
    name: 'advanced-formatted-usage',
    description: 'should return 200 with an advanced formatting.',
    payload: {
      "id": 12345,
      "name": "John Doe",
      "childrens": [
        {
          "name": "Jane",
          "age": 5
        },
        {
          "name": "Bob",
          "age": 8
        }
      ],
      "pets": [],
      "favoriteColors": {
        "primary": null,
        "secondary": "blue"
      },
      "lastLogin": "2023-06-28T18:30:00Z",
      "notes": null
    },
    expected: {
      user: {id: 12345, name: 'John Doe'},
      hasNotes: false,
      hasChildrens: true,
      childrenNames: ['Jane', 'Bob'],
      hasPets: false,
      favoriteColor: 'blue',
    },
    expectedResponse: ''
  },
]



const testSuite = () => {
  scenarios.forEach((test) => {
    describe(`${test.description} [${test.name}]`, async () => {
      const res = session(test.name).post('', JSON.stringify(test.payload), test.configuration);
      expect(res.status).to.equal(200);

      const storedValue = await redisClient.lpop(`integration:${test.name}`);
      console.log(`[${test.name}]`, storedValue);

      expect(JSON.parse(storedValue)).to.deep.equal(test.expected);
      expect(res.body).to.equal(test.expectedResponse)
    })
  });
}

export default testSuite;
