<?php
class InfoController
{
    public function insert(): void
    {
        header('Content-Type: application/json');
        $body = json_decode(file_get_contents('php://input'), true);

        // Demo: cố tình INSERT vào cột 'age' không tồn tại
        $pdo  = $this->getDb();
        $stmt = $pdo->prepare("INSERT INTO info (name, age) VALUES (?, ?)");
        $stmt->execute([$body['name'] ?? '', $body['age'] ?? 0]);

        echo json_encode(['status' => 'ok']);
    }

    public function getUser(): void
    {
        header('Content-Type: application/json');
        $pdo  = $this->getDb();
        $rows = $pdo->query("SELECT * FROM info")->fetchAll();
        echo json_encode($rows);
    }

    public function update(): void
    {
        header('Content-Type: application/json');
        $body = json_decode(file_get_contents('php://input'), true);
        $pdo  = $this->getDb();
        $stmt = $pdo->prepare("UPDATE info SET name = ? WHERE id = ?");
        $stmt->execute([$body['name'] ?? '', $body['id'] ?? 0]);
        echo json_encode(['status' => 'ok']);
    }

    private function getDb(): PDO
    {
        return new PDO(
            'mysql:host=' . ($_ENV['DB_HOST'] ?? 'mysql') . ';dbname=' . ($_ENV['DB_NAME'] ?? 'demo'),
            $_ENV['DB_USER'] ?? 'root',
            $_ENV['DB_PASS'] ?? 'root',
        );
    }
}