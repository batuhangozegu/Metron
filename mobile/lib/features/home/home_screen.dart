import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../data/providers.dart';
import '../../l10n/app_strings.dart';
import '../../l10n/live_strings.dart';
import '../../l10n/number_formatter.dart';
import '../../router.dart';
import '../../widgets/widgets.dart';
import '../../widgets/account_gate.dart';
import '../../widgets/live_state.dart';

class HomeScreen extends ConsumerWidget {
  const HomeScreen({super.key});
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final s = LiveStrings.of(context), f = NumberFormatter.of(context);
    return WarRoomScaffold(
      appBar: AppBar(title: CommandTitle(AppStrings.of(context).risk)),
      body: AccountGate(
        child: VisibleRefresh(
          refresh: () => refreshPortfolio(ref),
          child: ListView(
            padding: const EdgeInsets.all(16),
            children: [
              AsyncPanel(
                value: ref.watch(portfolioProvider),
                retry: () => ref.invalidate(portfolioProvider),
                data: (portfolio) {
                  if (portfolio == null) return Text(s.noAccounts);
                  final summary = portfolio['summary'] as Map<String, dynamic>;
                  final drawdown =
                      portfolio['max_drawdown'] as Map<String, dynamic>?;
                  return MetronCard(
                    child: Column(
                      children: [
                        ValueRow(
                          s.equity,
                          '${f.decimal(summary['equity'] as String)} USDT',
                        ),
                        ValueRow(
                          s.cash,
                          '${f.decimal(summary['cash'] as String)} USDT',
                        ),
                        ValueRow(
                          s.returns,
                          f.unit(
                            f.decimal(summary['return_percent'] as String),
                            '%',
                          ),
                        ),
                        ValueRow(
                          s.drawdown,
                          drawdown == null
                              ? '—'
                              : f.unit(
                                  f.decimal(drawdown['percent'] as String),
                                  '%',
                                ),
                        ),
                      ],
                    ),
                  );
                },
              ),
              const SizedBox(height: 20),
              SectionHeader(s.positions),
              const SizedBox(height: 12),
              AsyncPanel(
                value: ref.watch(positionsProvider),
                retry: () => ref.invalidate(positionsProvider),
                data: (positions) => positions.isEmpty
                    ? Text(s.emptyPositions)
                    : MetronCard(
                        padding: EdgeInsets.zero,
                        child: Column(
                          children: [
                            for (final p in positions)
                              ListTile(
                                title: Text(p.symbol),
                                subtitle: Text(
                                  '${p.side} · ${p.value('quantity')}',
                                ),
                                trailing: Num(
                                  f.decimal(p.value('unrealized_pnl')),
                                ),
                                onTap: () => context.pushNamed(
                                  AppRoutes.position,
                                  pathParameters: {'id': p.id},
                                ),
                              ),
                          ],
                        ),
                      ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
