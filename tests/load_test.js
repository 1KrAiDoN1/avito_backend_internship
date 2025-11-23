// Нагрузочное тестирование с использованием k6
// Установка: brew install k6 (macOS) или https://k6.io/docs/getting-started/installation/
// Запуск: k6 run tests/load_test.js

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Counter } from 'k6/metrics';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

// Метрики
const errorRate = new Rate('errors');
const realErrors = new Rate('real_errors');           // Реальные ошибки (5xx, неожиданные 4xx)
const expected404 = new Counter('expected_404s');     // Ожидаемые 404
const unexpected4xx = new Counter('unexpected_4xx');  // Неожиданные 4xx ошибки
const httpReqFailedNo404 = new Rate('http_req_failed_no_404'); // Ошибки без учета 404

// Конфигурация нагрузки
export const options = {
  stages: [
    { duration: '20s', target: 5 },   // Разогрев до 5 RPS
    { duration: '30s', target: 5 },   // Держим 5 RPS
    { duration: '20s', target: 10 },  // Увеличиваем до 10 RPS
    { duration: '20s', target: 10 }, // Держим 10 RPS
  ],
  thresholds: {
    http_req_duration: ['p(95)<300'],     // 95% запросов < 300ms (SLI)
    http_req_failed_no_404: ['rate<0.01'], // < 1% ошибок (без учета 404)
    real_errors: ['rate<0.01'],           // < 1% реальных ошибок (5xx, неожиданные 4xx)
    errors: ['rate<0.01'],                // < 1% ошибок в бизнес-логике
  },
};

// Тестовые данные
const teams = [
  {
    team_name: 'backend',
    members: [
      { user_id: 'alice', username: 'Alice', is_active: true },
      { user_id: 'bob', username: 'Bob', is_active: true },
    ],
  },
  {
    team_name: 'frontend',
    members: [
      { user_id: 'charlie', username: 'Charlie', is_active: true },
      { user_id: 'david', username: 'David', is_active: true },
    ],
  },
];

// Массив для хранения созданных PR
let createdPRs = [];

// Функция для классификации HTTP ответов
function classifyResponse(res, expectedStatuses = [200, 201]) {
  const status = res.status;
  
  // Успешные ответы
  if (expectedStatuses.includes(status)) {
    httpReqFailedNo404.add(0); // Успешный запрос
    return { isError: false, isRealError: false, is404: false };
  }
  
  // Ожидаемые 404 (не ошибка, а нормальное поведение API)
  if (status === 404) {
    expected404.add(1);
    httpReqFailedNo404.add(0); // 404 не считается ошибкой для этой метрики
    return { isError: false, isRealError: false, is404: true };
  }
  
  // Неожиданные 4xx ошибки (например, 400 bad request)
  if (status >= 400 && status < 500) {
    unexpected4xx.add(1);
    realErrors.add(1);
    httpReqFailedNo404.add(1); // Считаем как ошибку (кроме 404)
    return { isError: true, isRealError: true, is404: false };
  }
  
  // Серверные ошибки 5xx - это всегда реальная проблема
  if (status >= 500) {
    realErrors.add(1);
    httpReqFailedNo404.add(1); // Считаем как ошибку
    return { isError: true, isRealError: true, is404: false };
  }
  
  httpReqFailedNo404.add(0); // Другие статусы (например, 3xx) не считаются ошибками
  return { isError: false, isRealError: false, is404: false };
}

export function setup() {
  // Создаем тестовые команды
  teams.forEach(team => {
    const res = http.post(
      `${BASE_URL}/api/v1/team/add`,
      JSON.stringify(team),
      {
        headers: { 'Content-Type': 'application/json' },
      }
    );
    console.log(`Created team ${team.team_name}: ${res.status}`);
  });

  return { teams };
}

export default function (data) {
  const scenarios = [
    testCreatePR,
    testGetTeam,
    testSetUserActive,
    testGetUserReviews,
    testMergePR,
    testReassignReviewer,
    testGetStatistics,
  ];

  // Выбираем случайный сценарий
  const scenario = scenarios[Math.floor(Math.random() * scenarios.length)];
  scenario(data);

  sleep(1); // Пауза между итерациями
}

function testCreatePR(data) {
  const team = data.teams[Math.floor(Math.random() * data.teams.length)];
  const author = team.members[Math.floor(Math.random() * team.members.length)];
  const prId = `pr_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;

  const payload = {
    pull_request_id: prId,
    pull_request_name: `Test PR ${prId}`,
    author_id: author.user_id,
  };

  const res = http.post(
    `${BASE_URL}/api/v1/pullRequests/create`,
    JSON.stringify(payload),
    {
      headers: { 'Content-Type': 'application/json' },
      tags: { name: 'CreatePR' },
    }
  );

  const classification = classifyResponse(res, [201]);
  
  const success = check(res, {
    'PR created': (r) => r.status === 201,
    'response time < 300ms': (r) => r.timings.duration < 300,
  });

  if (res.status === 201) {
    createdPRs.push(prId);
    // Ограничиваем размер массива для экономии памяти
    if (createdPRs.length > 100) {
      createdPRs.shift();
    }
  }

  if (classification.isRealError) {
    errorRate.add(1);
  } else {
    errorRate.add(!success);
  }
}

function testGetTeam(data) {
  const team = data.teams[Math.floor(Math.random() * data.teams.length)];

  const res = http.get(
    `${BASE_URL}/api/v1/team/get?team_name=${team.team_name}`,
    {
      tags: { name: 'GetTeam' },
    }
  );

  const classification = classifyResponse(res, [200]);
  
  const success = check(res, {
    'team retrieved': (r) => r.status === 200,
    'response time < 300ms': (r) => r.timings.duration < 300,
  });

  if (classification.isRealError) {
    errorRate.add(1);
  } else {
    errorRate.add(!success);
  }
}

function testSetUserActive(data) {
  const team = data.teams[Math.floor(Math.random() * data.teams.length)];
  const user = team.members[Math.floor(Math.random() * team.members.length)];

  const payload = {
    user_id: user.user_id,
    is_active: Math.random() > 0.5,
  };

  const res = http.post(
    `${BASE_URL}/api/v1/users/setIsActive`,
    JSON.stringify(payload),
    {
      headers: { 'Content-Type': 'application/json' },
      tags: { name: 'SetUserActive' },
    }
  );

  const classification = classifyResponse(res, [200, 404]);
  
  const success = check(res, {
    'user updated': (r) => r.status === 200 || r.status === 404,
    'response time < 300ms': (r) => r.timings.duration < 300,
  });

  if (classification.isRealError) {
    errorRate.add(1);
  } else {
    errorRate.add(!success);
  }
}

function testGetUserReviews(data) {
  const team = data.teams[Math.floor(Math.random() * data.teams.length)];
  const user = team.members[Math.floor(Math.random() * team.members.length)];

  const res = http.get(
    `${BASE_URL}/api/v1/users/getReview?user_id=${user.user_id}`,
    {
      tags: { name: 'GetUserReviews' },
    }
  );

  const classification = classifyResponse(res, [200]);
  
  const success = check(res, {
    'reviews retrieved': (r) => r.status === 200,
    'response time < 300ms': (r) => r.timings.duration < 300,
  });

  if (classification.isRealError) {
    errorRate.add(1);
  } else {
    errorRate.add(!success);
  }
}

function testMergePR(data) {
  let prId;
  let expectingReal = false;
  
  // 90% времени используем реальный PR, 10% - несуществующий (для тестирования 404)
  if (createdPRs.length > 0 && Math.random() > 0.1) {
    prId = createdPRs[Math.floor(Math.random() * createdPRs.length)];
    expectingReal = true;
  } else {
    prId = `pr_nonexistent_${Math.random().toString(36).substr(2, 9)}`;
  }

  const payload = {
    pull_request_id: prId,
  };

  const res = http.post(
    `${BASE_URL}/api/v1/pullRequests/merge`,
    JSON.stringify(payload),
    {
      headers: { 'Content-Type': 'application/json' },
      tags: { name: 'MergePR' },
    }
  );

  const classification = classifyResponse(res, [200, 404, 400]);
  
  const success = check(res, {
    'PR merge attempted': (r) => r.status === 200 || r.status === 404 || r.status === 400,
    'response time < 300ms': (r) => r.timings.duration < 300,
  });

  // Если успешно смержили, удаляем из массива
  if (res.status === 200 && expectingReal) {
    const index = createdPRs.indexOf(prId);
    if (index > -1) {
      createdPRs.splice(index, 1);
    }
  }

  if (classification.isRealError) {
    errorRate.add(1);
  } else {
    errorRate.add(!success);
  }
}

function testReassignReviewer(data) {
  const team = data.teams[Math.floor(Math.random() * data.teams.length)];
  const user = team.members[Math.floor(Math.random() * team.members.length)];
  
  let prId;
  
  // 90% времени используем реальный PR, 10% - несуществующий (для тестирования 404)
  if (createdPRs.length > 0 && Math.random() > 0.1) {
    prId = createdPRs[Math.floor(Math.random() * createdPRs.length)];
  } else {
    prId = `pr_nonexistent_${Math.random().toString(36).substr(2, 9)}`;
  }

  const payload = {
    pull_request_id: prId,
    old_user_id: user.user_id,
  };

  const res = http.post(
    `${BASE_URL}/api/v1/pullRequests/reassign`,
    JSON.stringify(payload),
    {
      headers: { 'Content-Type': 'application/json' },
      tags: { name: 'ReassignReviewer' },
    }
  );

  const classification = classifyResponse(res, [200, 404, 400, 409]);
  
  const success = check(res, {
    'reassign attempted': (r) => r.status === 200 || r.status === 404 || r.status === 400 || r.status === 409,
    'response time < 300ms': (r) => r.timings.duration < 300,
  });

  if (classification.isRealError) {
    errorRate.add(1);
  } else {
    errorRate.add(!success);
  }
}

function testGetStatistics(data) {
  const res = http.get(
    `${BASE_URL}/api/v1/statistics`,
    {
      tags: { name: 'GetStatistics' },
    }
  );

  const classification = classifyResponse(res, [200]);
  
  const success = check(res, {
    'statistics retrieved': (r) => r.status === 200,
    'response time < 300ms': (r) => r.timings.duration < 300,
    'has data': (r) => {
      try {
        const stats = JSON.parse(r.body);
        return stats.assignments_by_user !== undefined && stats.pull_requests !== undefined;
      } catch (e) {
        return false;
      }
    },
  });

  if (classification.isRealError) {
    errorRate.add(1);
  } else {
    errorRate.add(!success);
  }
}

export function teardown(data) {
  console.log('Load test completed!');
  console.log('=====================================');
  console.log('Error Statistics Summary:');
  console.log('- 404 errors are counted separately (see "expected_404s" metric)');
  console.log('- http_req_failed_no_404: errors excluding 404 responses');
  console.log('- Real errors include: 5xx responses and unexpected 4xx codes');
  console.log('=====================================');
}

