import 'dart:convert';
import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:metron/app.dart';
import 'package:metron/data/api_client.dart';
import 'package:metron/data/providers.dart';
import 'package:metron/l10n/locale_provider.dart';

class MemoryStore implements PrivateStore {
  final values = <String, String>{};
  @override
  Future<String?> read(String key) async => values[key];
  @override
  Future<void> write(String key, String value) async {
    values[key] = value;
  }

  @override
  Future<void> delete(String key) async {
    values.remove(key);
  }
}

const testUser = {
  'id': 'user-1',
  'email': 'me@example.com',
  'two_factor_enabled': true,
  'created_at': '2026-11-23T00:00:00Z',
};
const accountJson = {
  'id': 'account-1',
  'type': 'paper',
  'name': 'Practice',
  'market': 'spot',
  'paper_balance': '10000',
};
const positionJson = {
  'id': 'position-1',
  'symbol': 'BTCUSDT',
  'side': 'long',
  'quantity': '0.00000123',
  'entry_price': '100.12345678',
  'stop': '95',
  'initial_risk': '10',
  'status': 'open',
  'last_price': '101',
  'unrealized_pnl': '0.12345678',
};
Json checkJson({bool blocked = false}) => {
  'symbol': 'BTCUSDT',
  'side': 'long',
  'market': 'spot',
  'leverage': 1,
  'allowed': !blocked,
  'verdict': blocked ? 'fail' : 'pass',
  'entry_price': '100.12345678',
  'stop': '95',
  'suggested_stop': '94.12345678',
  'target_price': '110',
  'position_size': '0.00000123',
  'notional': '0.0001231518518394',
  'actual_risk': '0.0000063018518394',
  'checks': [
    {
      'rule_id': blocked ? 'daily_loss' : 'atr_stop',
      'status': blocked ? 'fail' : 'pass',
      'message_en': blocked
          ? 'Daily loss reached; stop trading.'
          : 'Stop follows ATR.',
      'message_tr': blocked
          ? 'Günlük zarar sınırı doldu.'
          : 'Stop ATR ile uyumlu.',
      'values': {'current': '1', 'limit': '5'},
    },
  ],
};

class FakeBackend {
  Completer<void>? sessionWait;
  bool sessionFail = false, emptyMarkets = false;
  bool signedIn = true,
      twoFactor = false,
      blocked = false,
      rejectOrder = false,
      closed = false;
  bool marketsFail = false,
      portfolioFail = false,
      accountsFail = false,
      codeFails = false;
  final accounts = <Json>[
    {...accountJson},
  ];
  final requests = <http.Request>[];
  final store = MemoryStore();
  late final api = ApiClient(
    client: MockClient(handle),
    store: store,
    baseUrl: 'http://metron.test',
  );
  http.Response response(Object body, [int status = 200, String? cookie]) =>
      http.Response(
        jsonEncode(body),
        status,
        headers: {'content-type': 'application/json', 'set-cookie': ?cookie},
      );
  Future<http.Response> handle(http.Request request) async {
    requests.add(request);
    final path = request.url.path;
    if (path == '/v1/me') {
      await sessionWait?.future;
      if (sessionFail) return response({'error': 'unavailable'}, 503);
      return signedIn
          ? response(testUser)
          : response({'error': 'authentication required'}, 401);
    }
    if (path == '/v1/auth/login' || path == '/v1/auth/register') {
      if (twoFactor && path.endsWith('login')) {
        return response(
          {'two_factor_required': true},
          200,
          'metron_pending=pending; Path=/; Max-Age=300; HttpOnly',
        );
      }
      signedIn = true;
      return response(
        testUser,
        path.endsWith('register') ? 201 : 200,
        'metron_session=session; Max-Age=2592000; HttpOnly',
      );
    }
    if (path == '/v1/auth/login/2fa') {
      if (codeFails) return response({'error': 'invalid code'}, 401);
      signedIn = true;
      return response(
        testUser,
        200,
        'metron_session=full; Max-Age=2592000, metron_pending=; Max-Age=0',
      );
    }
    if (path == '/v1/auth/logout') {
      signedIn = false;
      return response({'ok': true});
    }
    if (path == '/v1/accounts') {
      if (accountsFail) return response({'error': 'unavailable'}, 503);
      if (request.method == 'POST') {
        final body = jsonDecode(request.body) as Json;
        final result = {...accountJson, ...body};
        accounts.add(result);
        return response(result, 201);
      }
      return response(accounts);
    }
    if (path == '/v1/markets') {
      if (marketsFail) return response({'error': 'unavailable'}, 503);
      return response({
        'markets': emptyMarkets
            ? []
            : [
                for (final symbol in [
                  'BTC',
                  'ETH',
                  'SOL',
                  'AVAX',
                  'DOT',
                  'ADA',
                  'LINK',
                ])
                  {
                    'symbol': '${symbol}USDT',
                    'base_asset': symbol,
                    'quote_asset': 'USDT',
                    'coin_class': 'core',
                    'price': '100.12345678',
                    'change_24h_percent': '-1.25',
                    'atr_percent': '2.25',
                    'suggested_stop_percent': '3.375',
                  },
              ],
      });
    }
    if (path.endsWith('/portfolio')) {
      if (portfolioFail) {
        return response({'error': 'portfolio_unavailable'}, 503);
      }
      return response({
        'summary': {
          'equity': '10000.12345678',
          'cash': '9000.12',
          'return_percent': '1.25',
        },
        'max_drawdown': {'percent': '2.5'},
      });
    }
    if (path.endsWith('/positions')) {
      return response({
        'positions': closed ? [] : [positionJson],
      });
    }
    if (path.endsWith('/close')) {
      closed = true;
      return response({
        'position': {...positionJson, 'status': 'closed'},
      });
    }
    if (path.endsWith('/risk/check')) {
      return response(checkJson(blocked: blocked));
    }
    if (path == '/v1/orders') {
      if (rejectOrder) {
        return response({
          'error': {
            'id': 'risk_blocked',
            'message_en': 'Blocked',
            'message_tr': 'Engellendi',
          },
          'check': checkJson(blocked: true),
        }, 422);
      }
      return response({'position': positionJson, 'check': checkJson()}, 201);
    }
    if (path.endsWith('/risk-profile')) {
      return response(
        request.method == 'PUT'
            ? jsonDecode(request.body)
            : {
                'risk_per_trade_percent': '2',
                'target_risk_reward': '2',
                'total_open_risk_limit_percent': '6',
                'max_open_positions': 5,
                'max_losing_trades_today': 3,
                'daily_loss_limit_percent': '5',
                'max_stopped_trades_this_week': 5,
                'weekly_loss_limit_percent': '10',
                'updated_at': null,
              },
      );
    }
    return response({'error': 'not found'}, 404);
  }
}

Future<ProviderContainer> bootLive(
  WidgetTester tester, {
  FakeBackend? backend,
  String language = 'en',
  bool light = false,
  bool settle = true,
}) async {
  final fake = backend ?? FakeBackend();
  GoogleFonts.config.allowRuntimeFetching = false;
  tester.view.physicalSize = const Size(390, 1000);
  tester.view.devicePixelRatio = 1;
  addTearDown(tester.view.resetPhysicalSize);
  addTearDown(tester.view.resetDevicePixelRatio);
  final container = ProviderContainer(
    retry: (_, _) => null,
    overrides: [apiClientProvider.overrideWithValue(fake.api)],
  );
  container.read(localeProvider.notifier).select(Locale(language));
  addTearDown(() async {
    await tester.pumpWidget(const SizedBox.shrink());
    container.dispose();
    fake.api.close();
  });
  await tester.pumpWidget(
    UncontrolledProviderScope(
      container: container,
      child: MetronApp(themeMode: light ? ThemeMode.light : ThemeMode.dark),
    ),
  );
  await tester.runAsync(() => GoogleFonts.pendingFonts());
  if (settle) {
    await tester.pumpAndSettle();
  } else {
    await tester.pump();
  }
  return container;
}
