import http from 'k6/http';
import { check, group, sleep } from 'k6';
import { TestRelationAPI } from './api/relation.js';
import { TestCycle } from './scenario/cycle.js';
import { TestRequiredAttr } from './scenario/required_attr.js';
import { TestReservedWord } from './scenario/reserved_word.js';

export const options = {
  vus: 1,
}

export default function() {
  const SERVER_URL = "http://localhost:8080"
  const Headers = {
    'Content-Type': 'application/json',
  }

  // healthy
  let res = http.get(`${SERVER_URL}/healthy`);
  check(res, { 'Server is healthy': (r) => r.status == 200 });

  res = http.post(`${SERVER_URL}/relation/clear-all-relations`, null, {headers:Headers});
  check(res, { 'ClearAllRelations': (r) => r.status == 200 });

  group("api", () => {
    group("relation", () => {
      TestRelationAPI(SERVER_URL, Headers);
    });
  });

  group("scenario", () => {
    group("cycle", () => {
      TestCycle(SERVER_URL, Headers);
    });
    group("require_attr", () => {
      TestRequiredAttr(SERVER_URL, Headers);
    });
    group("reserved_word", () => {
      TestReservedWord(SERVER_URL, Headers);
    });
  });
}
