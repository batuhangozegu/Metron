import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:google_fonts/google_fonts.dart';

import '../data/providers.dart';
import '../l10n/app_strings.dart';
import '../l10n/live_strings.dart';
import 'live_state.dart';
import '../l10n/number_formatter.dart';
import '../theme/metron_colors.dart';
import '../theme/tokens.dart';

/// Live quotes retain decimal strings; only animation geometry is numeric.
class MarketTicker extends ConsumerStatefulWidget {
  const MarketTicker({super.key});

  @override
  ConsumerState<MarketTicker> createState() => _MarketTickerState();
}

class _MarketTickerState extends ConsumerState<MarketTicker> {
  final _scroll = ScrollController();
  Timer? _timer;
  bool _paused = false;
  bool _reduceMotion = false;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    _reduceMotion =
        MediaQuery.disableAnimationsOf(context) ||
        MediaQuery.of(context).accessibleNavigation ||
        !TickerMode.valuesOf(context).enabled;
    _schedule();
  }

  void _schedule() {
    _timer?.cancel();
    if (_paused || _reduceMotion) return;
    _timer = Timer.periodic(const Duration(seconds: 4), (_) {
      if (!_scroll.hasClients || _scroll.position.isScrollingNotifier.value) {
        return;
      }
      final end = _scroll.position.maxScrollExtent;
      if (_scroll.offset >= end) {
        _scroll.jumpTo(0);
      } else {
        _scroll.animateTo(
          (_scroll.offset + 124).clamp(0, end),
          duration: const Duration(milliseconds: 1800),
          curve: Curves.linear,
        );
      }
    });
  }

  @override
  void dispose() {
    _timer?.cancel();
    _scroll.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final c = context.colors;
    final strings = AppStrings.of(context);
    final f = NumberFormatter.of(context);
    final feed = ref.watch(marketsProvider);
    final quotes = feed.value ?? [];
    final live = LiveStrings.of(context);
    return VisibleRefresh(
      refresh: () => ref.invalidate(marketsProvider),
      child: Semantics(
        label: strings.ticker,
        container: true,
        child: Container(
          decoration: BoxDecoration(
            color: c.surface,
            border: Border(bottom: BorderSide(color: c.limitEdge)),
          ),
          height: 44 * MediaQuery.textScalerOf(context).scale(1).clamp(1, 2),
          child: Row(
            children: [
              Container(
                padding: const EdgeInsets.only(left: 12),
                decoration: BoxDecoration(
                  color: c.raised,
                  border: Border(right: BorderSide(color: c.line)),
                ),
                child: Row(
                  children: [
                    Container(width: 5, height: 5, color: c.muted),
                    const SizedBox(width: 6),
                    Text(
                      feed.hasError
                          ? live.unavailable
                          : feed.isLoading
                          ? live.loading
                          : quotes.any((q) => q.price != null)
                          ? live.feed
                          : live.unavailable,
                      style: MetronType.label.copyWith(
                        fontSize: 10,
                        color: c.muted,
                        letterSpacing: .7,
                      ),
                    ),
                    IconButton(
                      tooltip: _paused
                          ? strings.resumeTicker
                          : strings.pauseTicker,
                      isSelected: _paused,
                      iconSize: 16,
                      onPressed: () {
                        setState(() => _paused = !_paused);
                        // Stop an in-flight scroll immediately as well.
                        if (_scroll.hasClients) _scroll.jumpTo(_scroll.offset);
                        _schedule();
                      },
                      icon: Icon(_paused ? Icons.play_arrow : Icons.pause),
                    ),
                  ],
                ),
              ),
              Expanded(
                child: ListView.builder(
                  key: const ValueKey('ticker-quotes'),
                  controller: _scroll,
                  scrollDirection: Axis.horizontal,
                  itemCount: quotes.length,
                  itemBuilder: (context, index) {
                    final quote = quotes[index];
                    final change = quote.change;
                    final price = f.decimal(quote.price ?? '—');
                    final changeText = change == null
                        ? '—'
                        : f.unit(f.decimal(change), '%');
                    return Semantics(
                      label:
                          '${quote.pair}, $price, ${strings.dailyChange} $changeText',
                      excludeSemantics: true,
                      child: Container(
                        alignment: Alignment.center,
                        padding: const EdgeInsets.symmetric(horizontal: 16),
                        decoration: BoxDecoration(
                          border: Border(right: BorderSide(color: c.line)),
                        ),
                        child: Text.rich(
                          TextSpan(
                            children: [
                              TextSpan(
                                text: '${quote.pair}   ',
                                style: TextStyle(
                                  color: c.text,
                                  fontSize: 11,
                                  fontWeight: FontWeight.w600,
                                ),
                              ),
                              TextSpan(
                                text: '$price   ',
                                style: GoogleFonts.jetBrainsMono(
                                  textStyle: MetronType.figure.copyWith(
                                    fontSize: 12,
                                    color: c.text,
                                  ),
                                ),
                              ),
                              TextSpan(
                                text: changeText,
                                style: GoogleFonts.jetBrainsMono(
                                  textStyle: MetronType.figure.copyWith(
                                    fontSize: 11,
                                    color: change?.startsWith('-') == true
                                        ? c.loss
                                        : change != null && change != '0'
                                        ? c.profit
                                        : c.muted,
                                  ),
                                ),
                              ),
                            ],
                          ),
                        ),
                      ),
                    );
                  },
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
