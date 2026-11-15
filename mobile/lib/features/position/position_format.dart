/// Group decimal fixture strings without doing financial arithmetic.
String formatDecimal(String value) {
  final parts = value.split('.');
  final grouped = parts.first.replaceAllMapped(
    RegExp(r'(\d)(?=(\d{3})+(?!\d))'),
    (m) => '${m[1]},',
  );
  return parts.length == 1 || parts[1] == '0'
      ? grouped
      : '$grouped.${parts[1]}';
}
