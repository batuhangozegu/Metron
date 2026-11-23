import 'dart:async';
import 'dart:convert';

import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:metron/data/api_client.dart';

import 'live_test_app.dart';

void main() {
  test(
    'Cookies persist, refresh, restore, send, and clear on unauthorized',
    () async {
      final store = MemoryStore();
      var step = 0;
      final api = ApiClient(
        store: store,
        baseUrl: 'http://example.test',
        client: MockClient((request) async {
          step++;
          if (step == 1) {
            return http.Response(
              '{}',
              200,
              headers: {
                'set-cookie': 'metron_session=first; Expires=Wed, 25 Nov 2026 12:00:00 GMT; Max-Age=2592000; HttpOnly, metron_pending=; Max-Age=0',
              },
            );
          }
          expect(request.headers['cookie'], 'metron_session=first');
          return http.Response(
            '{}',
            200,
            headers: {
              'set-cookie':
                  'metron_session=refreshed; Max-Age=2592000; HttpOnly',
            },
          );
        }),
      );
      await api.request(
        '/auth/login',
        body: {'paper_balance': '10000.12345678'},
      );
      expect(jsonDecode(store.values[api.sessionKey]!)['value'], 'first');
      await api.request('/me');
      expect(jsonDecode(store.values[api.sessionKey]!)['value'], 'refreshed');
      var unauthorized = false;
      final restored = ApiClient(
        store: store,
        baseUrl: api.baseUrl,
        client: MockClient((request) async {
          expect(request.headers['cookie'], 'metron_session=refreshed');
          return http.Response('{"error":"authentication required"}', 401);
        }),
      )..onUnauthorized = () => unauthorized = true;
      await expectLater(
        restored.request('/accounts'),
        throwsA(isA<ApiError>().having((e) => e.status, 'status', 401)),
      );
      expect(unauthorized, isTrue);
      expect(store.values[api.sessionKey], isNull);
      api.close();
      restored.close();
    },
  );
  test('Pending cookie survives invalid 2FA, is not persisted and is deleted after verification', () async {
    final store = MemoryStore();
    var step = 0;
    final api = ApiClient(
      store: store,
      client: MockClient((r) async {
        step++;
        if (step == 1) {
          return http.Response(
            '{}',
            200,
            headers: {'set-cookie': 'metron_pending=temporary; Max-Age=300'},
          );
        }
        expect(r.headers['cookie'], 'metron_pending=temporary');
        if (step == 2) return http.Response('{}', 401);
        return http.Response(
          '{}',
          200,
          headers: {
            'set-cookie': 'metron_session=full; Max-Age=2592000, metron_pending=; Max-Age=-1',
          },
        );
      }),
    );
    await api.request('/auth/login', body: {});
    expect(store.values, isEmpty);
    await expectLater(
      api.request('/auth/login/2fa', body: {'code': '000000'}),
      throwsA(isA<ApiError>()),
    );
    await api.request('/auth/login/2fa', body: {'code': '654321'});
    expect(store.values.values.single, contains('full'));
    api.close();
  });
  test('Logout clears cookies even when the service fails', () async {
    final store = MemoryStore();
    var step = 0;
    final api = ApiClient(
      store: store,
      client: MockClient((r) async {
        step++;
        if (step == 1) {
          return http.Response(
            '{}',
            200,
            headers: {'set-cookie': 'metron_session=full; Max-Age=2592000'},
          );
        }
        expect(r.method, 'POST');
        expect(r.headers['content-type'], 'application/json');
        return http.Response('{}', 503);
      }),
    );
    await api.request('/auth/login', body: {});
    await expectLater(api.logout(), throwsA(isA<ApiError>()));
    expect(store.values, isEmpty);
    api.close();
  });
  test('Late refresh cannot revive a cleared session', () async {
    final completer = Completer<http.Response>();
    final store = MemoryStore();
    final api = ApiClient(
      store: store,
      client: MockClient((r) => completer.future),
    );
    final pending = api.request('/accounts');
    await Future<void>.delayed(Duration.zero);
    await api.clearSession();
    completer.complete(
      http.Response(
        '{}',
        200,
        headers: {'set-cookie': 'metron_session=stale; Max-Age=2592000'},
      ),
    );
    await expectLater(pending, throwsA(isA<ApiError>()));
    expect(store.values, isEmpty);
    api.close();
  });
}
