<?php
class ExceptionHandler
{
    private static string $logDir = '/var/www/storage/logs';

    public static function register(): void
    {
        set_exception_handler([self::class, 'handle']);
        set_error_handler([self::class, 'handleError']);
    }

    public static function handle(Throwable $e): void
    {
        http_response_code(500);
        header('Content-Type: application/json');

        self::writeLog('ERROR', $e->getMessage(), [
            'exception' => get_class($e),
            'file'      => $e->getFile(),
            'line'      => $e->getLine(),
            'trace'     => self::formatTrace($e->getTrace()),
            'url'       => ($_SERVER['REQUEST_METHOD'] ?? '') . ' ' . ($_SERVER['REQUEST_URI'] ?? ''),
            'body'      => json_decode(file_get_contents('php://input'), true) ?? [],
        ]);

        echo json_encode(['error' => 'Internal Server Error']);
    }

    private static function writeLog(string $level, string $message, array $context): void
    {
        if (!is_dir(self::$logDir)) {
            mkdir(self::$logDir, 0755, true);
        }

        $file = self::$logDir . '/app-' . date('Y-m-d') . '.log';
        $ctx  = json_encode($context, JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES);
        $line = '[' . date('Y-m-d H:i:s') . '] app.' . $level . ': ' . $message . ' ' . $ctx . PHP_EOL;

        file_put_contents($file, $line, FILE_APPEND | LOCK_EX);
    }

    private static function formatTrace(array $trace): array
    {
        return array_map(fn($t) => [
            'file' => $t['file'] ?? '',
            'line' => $t['line'] ?? 0,
            'func' => ($t['class'] ?? '') . ($t['type'] ?? '') . ($t['function'] ?? ''),
        ], array_slice($trace, 0, 6));
    }

    public static function handleError(int $errno, string $errstr, string $errfile, int $errline): bool
    {
        self::handle(new ErrorException($errstr, $errno, $errno, $errfile, $errline));
        return true;
    }
}