import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../data/providers.dart';
import '../../l10n/app_strings.dart';
import '../../l10n/live_strings.dart';
import '../../l10n/number_formatter.dart';
import '../../router.dart';
import '../../widgets/widgets.dart';
import '../../widgets/live_state.dart';
import 'markets_strings.dart';

class MarketsScreen extends ConsumerStatefulWidget {
  const MarketsScreen({super.key});
  @override
  ConsumerState<MarketsScreen> createState() => _MarketsScreenState();
}

class _MarketsScreenState extends ConsumerState<MarketsScreen> {
  String _query = '';
  @override
  Widget build(BuildContext context) {
    final s = LiveStrings.of(context), f = NumberFormatter.of(context);
    return WarRoomScaffold(
      appBar: AppBar(title: CommandTitle(AppStrings.of(context).markets)),
      body: Column(
        children: [
          Padding(
            padding: const EdgeInsets.all(16),
            child: TextField(
              decoration: InputDecoration(
                labelText: MarketsStrings.of(context).search,
              ),
              onChanged: (v) => setState(() => _query = v),
            ),
          ),
          Expanded(
            child: AsyncPanel(
              value: ref.watch(marketsProvider),
              retry: () => ref.invalidate(marketsProvider),
              data: (markets) {
                final filtered = markets
                    .where(
                      (m) =>
                          m.pair.toLowerCase().contains(_query.toLowerCase()),
                    )
                    .toList();
                if (filtered.isEmpty) {
                  return Center(
                    child: Text(
                      markets.isEmpty
                          ? s.noMarkets
                          : MarketsStrings.of(context).noResults,
                    ),
                  );
                }
                return ListView(
                  children: [
                    for (final m in filtered)
                      ListTile(
                        title: Text(m.pair),
                        subtitle: Text(
                          '${m.coinClass} · ATR ${f.decimal(m.atr ?? '—')}% · ${MarketsStrings.of(context).stop} ${f.decimal(m.stop ?? '—')}%',
                        ),
                        trailing: Column(
                          mainAxisAlignment: MainAxisAlignment.center,
                          crossAxisAlignment: CrossAxisAlignment.end,
                          children: [
                            Num(f.decimal(m.price ?? '—')),
                            Num(f.unit(f.decimal(m.change ?? '—'), '%')),
                          ],
                        ),
                        onTap: () => context.pushNamed(
                          AppRoutes.check,
                          queryParameters: {'pair': m.pair},
                        ),
                      ),
                  ],
                );
              },
            ),
          ),
        ],
      ),
    );
  }
}
