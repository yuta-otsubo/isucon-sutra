<?php

declare(strict_types=1);

namespace IsuRide\Application\Database\Model;

readonly class Owner
{
    public function __construct(
        public string $id,
        public string $name,
        public string $accessToken,
        public int $createdAt,
        public int $updatedAt
    ) {
    }
}
