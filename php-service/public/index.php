<?php
require_once __DIR__ . '/../app/Core/ExceptionHandler.php';
require_once __DIR__ . '/../app/Core/Router.php';
require_once __DIR__ . '/../app/Controllers/InfoController.php';

ExceptionHandler::register();

$router = new Router();
$router->post('/insert-info', [InfoController::class, 'insert']);
$router->get('/get-user',     [InfoController::class, 'getUser']);
$router->put('/update-info',  [InfoController::class, 'update']);
$router->dispatch();