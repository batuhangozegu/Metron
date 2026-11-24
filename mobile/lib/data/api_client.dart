import 'dart:async';
import 'dart:convert';

import 'package:flutter/foundation.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:http/http.dart' as http;

typedef Json = Map<String, dynamic>;

abstract interface class PrivateStore {
  Future<String?> read(String key);
  Future<void> write(String key, String value);
  Future<void> delete(String key);
}

class SecureStore implements PrivateStore {
  const SecureStore();
  static const _storage = FlutterSecureStorage();
  @override
  Future<String?> read(String key) => _storage.read(key: key);
  @override
  Future<void> write(String key, String value) =>
      _storage.write(key: key, value: value);
  @override
  Future<void> delete(String key) => _storage.delete(key: key);
}

class ApiError implements Exception {
  const ApiError(this.status, this.data);
  final int status;
  final Json data;
  String? get id => data['error'] is Map
      ? (data['error'] as Map)['id'] as String?
      : data['code'] as String?;
}

/// Native cookie transport. The pending 2FA cookie stays in memory only.
class ApiClient {
  ApiClient({http.Client? client, PrivateStore? store, String? baseUrl})
    : _client = client ?? http.Client(),
      store = store ?? const SecureStore(),
      baseUrl = (baseUrl ?? defaultBaseUrl).replaceFirst(RegExp(r'/+$'), '');

  static String get defaultBaseUrl =>
      const String.fromEnvironment('METRON_API_URL').isNotEmpty
      ? const String.fromEnvironment('METRON_API_URL')
      : defaultTargetPlatform == TargetPlatform.android
      ? 'http://10.0.2.2:8080'
      : 'http://127.0.0.1:8080';
  final http.Client _client;
  final PrivateStore store;
  final String baseUrl;
  String get sessionKey => 'metron.session.$baseUrl';
  final Map<String, Json> _cookies = {};
  Future<void>? _loaded;
  Future<void> _storageQueue = Future.value();
  int _generation = 0;
  VoidCallback? onUnauthorized;

  Future<void> _persist() {
    final cookie = _cookies['metron_session'];
    final value = cookie == null ? null : jsonEncode(cookie);
    final operation = _storageQueue.then(
      (_) => value == null
          ? store.delete(sessionKey)
          : store.write(sessionKey, value),
    );
    _storageQueue = operation.catchError((Object _) {});
    return operation;
  }

  Future<void> _load() async {
    try {
      await (_loaded ??= _restore());
    } catch (_) {
      _loaded = null;
      rethrow;
    }
  }

  Future<void> _restore() async {
    final value = await store.read(sessionKey);
    if (value == null) return;
    try {
      final cookie = jsonDecode(value) as Json;
      if (cookie['value'] is String && _valid(cookie)) {
        _cookies['metron_session'] = cookie;
      } else {
        await store.delete(sessionKey);
      }
    } on FormatException {
      await store.delete(sessionKey);
    } on TypeError {
      await store.delete(sessionKey);
    }
  }

  bool _valid(Json cookie) =>
      DateTime.tryParse(cookie['expires'] as String? ?? '')
          ?.isAfter(DateTime.now().toUtc()) ??
      false;

  Future<void> clearSession({bool pending = false}) async {
    await _load();
    _generation++;
    _cookies.remove('metron_session');
    if (pending) _cookies.remove('metron_pending');
    await _persist();
  }

  Future<void> _capture(String? header) async {
    if (header == null) return;
    // http joins Set-Cookie values. Expires commas are not followed by name=.
    for (final line in header.split(RegExp(r',(?=\s*[^;,=\s]+=)'))) {
      final parts = line.split(';').map((s) => s.trim()).toList();
      final split = parts.first.indexOf('=');
      if (split < 1) continue;
      final name = parts.first.substring(0, split);
      if (name != 'metron_session' && name != 'metron_pending') continue;
      final value = parts.first.substring(split + 1);
      final age = parts
          .where((s) => s.toLowerCase().startsWith('max-age='))
          .map((s) => int.tryParse(s.substring(8)))
          .firstOrNull;
      if (value.isEmpty || (age != null && age <= 0)) {
        _cookies.remove(name);
      } else {
        _cookies[name] = {
          'value': value,
          'expires': DateTime.now()
              .toUtc()
              .add(
                Duration(
                  seconds: age ?? (name == 'metron_pending' ? 300 : 2592000),
                ),
              )
              .toIso8601String(),
          'secure': parts.any((s) => s.toLowerCase() == 'secure'),
        };
      }
    }
    await _persist();
  }

  Future<dynamic> request(String path, {Json? body, String? method}) async {
    await _load();
    final generation = _generation;
    final uri = Uri.parse('$baseUrl/v1$path');
    final request = http.Request(
      method ?? (body == null ? 'GET' : 'POST'),
      uri,
    );
    request.followRedirects = false;
    request.headers['Accept'] = 'application/json';
    final cookie = _cookies.entries
        .where(
          (e) =>
              _valid(e.value) &&
              (e.value['secure'] != true || uri.scheme == 'https'),
        )
        .map((e) => '${e.key}=${e.value['value']}')
        .join('; ');
    if (cookie.isNotEmpty) request.headers['Cookie'] = cookie;
    if (request.method != 'GET') {
      request.headers['Content-Type'] = 'application/json';
      request.body = jsonEncode(body ?? <String, dynamic>{});
    }
    late http.Response response;
    try {
      response = await (() async => http.Response.fromStream(
        await _client.send(request),
      ))().timeout(const Duration(seconds: 15));
    } on TimeoutException {
      throw const ApiError(0, {});
    } on http.ClientException {
      throw const ApiError(0, {});
    }
    if (generation != _generation) throw const ApiError(401, {});
    await _capture(response.headers['set-cookie']);
    if (response.statusCode == 401) {
      await clearSession();
      if (!path.startsWith('/auth/')) onUnauthorized?.call();
    }
    if (response.statusCode == 204) return null;
    dynamic data;
    try {
      data = jsonDecode(utf8.decode(response.bodyBytes));
    } on FormatException {
      throw ApiError(
        response.statusCode >= 400 ? response.statusCode : 502,
        {},
      );
    }
    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw ApiError(response.statusCode, data is Json ? data : {});
    }
    return data;
  }

  Future<void> logout() async {
    try {
      await request('/auth/logout', body: {});
    } finally {
      await clearSession(pending: true);
    }
  }

  void close() => _client.close();
}
