import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'api_client.dart';
import 'live_models.dart';

final apiClientProvider = Provider<ApiClient>((ref) {
  final client = ApiClient();
  ref.onDispose(client.close);
  return client;
});
final sessionProvider = AsyncNotifierProvider<Session, Json?>(Session.new);

class Session extends AsyncNotifier<Json?> {
  @override
  Future<Json?> build() async {
    final api = ref.watch(apiClientProvider);
    api.onUnauthorized = () {
      state = const AsyncData(null);
    };
    ref.onDispose(() => api.onUnauthorized = null);
    try {
      return await api.request('/me') as Json;
    } on ApiError catch (e) {
      if (e.status == 401) return null;
      rethrow;
    }
  }

  Future<bool> login(
    String email,
    String password, {
    bool register = false,
  }) async {
    final result =
        await ref
                .read(apiClientProvider)
                .request(
                  register ? '/auth/register' : '/auth/login',
                  body: {'email': email.trim(), 'password': password},
                )
            as Json;
    if (result['two_factor_required'] == true) return false;
    state = AsyncData(result);
    return true;
  }

  Future<void> complete(String code) async {
    final user =
        await ref
                .read(apiClientProvider)
                .request('/auth/login/2fa', body: {'code': code})
            as Json;
    state = AsyncData(user);
  }

  Future<void> logout() async {
    try {
      await ref.read(apiClientProvider).logout();
    } finally {
      state = const AsyncData(null);
    }
  }
}

final accountsProvider = AsyncNotifierProvider<Accounts, List<TradingAccount>>(
  Accounts.new,
);
final selectedAccountIdProvider = NotifierProvider<SelectedAccountId, String?>(
  SelectedAccountId.new,
);

class SelectedAccountId extends Notifier<String?> {
  @override
  String? build() => null;
  void select(String? value) => state = value;
}

class Accounts extends AsyncNotifier<List<TradingAccount>> {
  String _key(String userId) =>
      'metron.account.${ref.read(apiClientProvider).baseUrl}.$userId';
  @override
  Future<List<TradingAccount>> build() async {
    final user = await ref.watch(sessionProvider.future);
    if (user == null) return [];
    final api = ref.read(apiClientProvider);
    final data = await api.request('/accounts') as List;
    final accounts = data
        .where((a) => a['type'] == 'paper')
        .map((a) => TradingAccount.fromJson(a as Json))
        .toList();
    final remembered = await api.store.read(_key(user['id'] as String));
    if (!ref.mounted) return accounts;
    final selected =
        accounts.where((a) => a.id == remembered).firstOrNull ??
        accounts.firstOrNull;
    ref.read(selectedAccountIdProvider.notifier).select(selected?.id);
    return accounts;
  }

  Future<void> select(String id) async {
    final user = ref.read(sessionProvider).requireValue!;
    await ref
        .read(apiClientProvider)
        .store
        .write(_key(user['id'] as String), id);
    if (ref.mounted) ref.read(selectedAccountIdProvider.notifier).select(id);
  }

  Future<void> create(String name, String balance, String market) async {
    final result =
        await ref
                .read(apiClientProvider)
                .request(
                  '/accounts',
                  body: {
                    'name': name.trim(),
                    'paper_balance': balance,
                    'market': market,
                  },
                )
            as Json;
    final account = TradingAccount.fromJson(result);
    state = AsyncData([...state.requireValue, account]);
    await select(account.id);
  }
}

final selectedAccountProvider = Provider<TradingAccount?>((ref) {
  final id = ref.watch(selectedAccountIdProvider);
  return ref
      .watch(accountsProvider)
      .value
      ?.where((a) => a.id == id)
      .firstOrNull;
});
final marketsProvider = FutureProvider.autoDispose<List<MarketQuote>>((
  ref,
) async {
  final data = await ref.watch(apiClientProvider).request('/markets') as Json;
  return (data['markets'] as List)
      .map((m) => MarketQuote.fromJson(m as Json))
      .toList();
});
final portfolioProvider = FutureProvider.autoDispose<Json?>((ref) async {
  final account = ref.watch(selectedAccountProvider);
  if (account == null) return null;
  return await ref
          .watch(apiClientProvider)
          .request('/accounts/${account.id}/portfolio')
      as Json;
});
final positionsProvider = FutureProvider.autoDispose<List<LivePosition>>((
  ref,
) async {
  final account = ref.watch(selectedAccountProvider);
  if (account == null) return [];
  final data =
      await ref
              .watch(apiClientProvider)
              .request('/accounts/${account.id}/positions')
          as Json;
  return (data['positions'] as List)
      .map((p) => LivePosition.fromJson(p as Json))
      .toList();
});
final riskProfileProvider = FutureProvider.autoDispose<Json?>((ref) async {
  final account = ref.watch(selectedAccountProvider);
  if (account == null) return null;
  return await ref
          .watch(apiClientProvider)
          .request('/accounts/${account.id}/risk-profile')
      as Json;
});
void refreshPortfolio(WidgetRef ref) {
  ref.invalidate(portfolioProvider);
  ref.invalidate(positionsProvider);
}
