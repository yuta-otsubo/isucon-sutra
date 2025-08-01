<?php

declare(strict_types=1);

namespace IsuRide\Handlers\Owner;

use Fig\Http\Message\StatusCodeInterface;
use IsuRide\Database\Model\ChairWithDetail;
use IsuRide\Database\Model\Owner;
use IsuRide\Handlers\AbstractHttpHandler;
use IsuRide\Model\OwnerGetChairs200Response;
use IsuRide\Model\OwnerGetChairs200ResponseChairsInner;
use IsuRide\Response\ErrorResponse;
use PDO;
use PDOException;
use Psr\Http\Message\ResponseInterface;
use Psr\Http\Message\ServerRequestInterface;

class GetChairs extends AbstractHttpHandler
{
    public function __construct(
        private readonly PDO $db,
    ) {
    }

    /**
     * @param ServerRequestInterface $request
     * @param ResponseInterface $response
     * @param array<string, string> $args
     * @return ResponseInterface
     * @throws \Exception
     */
    public function __invoke(
        ServerRequestInterface $request,
        ResponseInterface $response,
        array $args
    ): ResponseInterface {
        /** @var Owner $owner */
        $owner = $request->getAttribute('owner');
        /** @var ChairWithDetail[] $chairs */
        $chairs = [];
        try {
            $stmt = $this->db->prepare(
                <<<SQL
SELECT id,
       owner_id,
       name,
       access_token,
       model,
       is_active,
       created_at,
       updated_at,
       IFNULL(total_distance, 0) AS total_distance,
       total_distance_updated_at
FROM chairs
       LEFT JOIN (
           SELECT chair_id,
                  SUM(IFNULL(distance, 0)) AS total_distance,
                  MAX(created_at) AS total_distance_updated_at
           FROM (
               SELECT chair_id,
                      created_at,
                      ABS(latitude - LAG(latitude) OVER (PARTITION BY chair_id ORDER BY created_at)) +
                      ABS(longitude - LAG(longitude) OVER (PARTITION BY chair_id ORDER BY created_at)) AS distance
               FROM chair_locations
           ) tmp
           GROUP BY chair_id
       ) distance_table ON distance_table.chair_id = chairs.id
WHERE owner_id = ?
SQL
            );
            $stmt->bindValue(1, $owner->id, PDO::PARAM_STR);
            $stmt->execute();
            $chairResult = $stmt->fetchAll(PDO::FETCH_ASSOC);
            foreach ($chairResult as $row) {
                $chairs[] = new ChairWithDetail(
                    id: $row['id'],
                    ownerId: $row['owner_id'],
                    name: $row['name'],
                    accessToken: $row['access_token'],
                    model: $row['model'],
                    isActive: (bool)$row['is_active'],
                    createdAt: $row['created_at'],
                    updatedAt: $row['updated_at'],
                    totalDistance: (int)$row['total_distance'],
                    totalDistanceUpdatedAt: $row['total_distance_updated_at']
                );
            }
        } catch (PDOException $e) {
            return (new ErrorResponse())->write(
                $response,
                StatusCodeInterface::STATUS_INTERNAL_SERVER_ERROR,
                $e
            );
        }
        $res = new OwnerGetChairs200Response();
        $ownerChairs = [];
        foreach ($chairs as $row) {
            $ownerChair = new OwnerGetChairs200ResponseChairsInner();
            $ownerChair->setId($row->id)
                ->setName($row->name)
                ->setModel($row->model)
                ->setActive($row->isActive)
                ->setRegisteredAt($row->createdAtUnixMilliseconds())
                ->setTotalDistance($row->totalDistance);
            if ($row->isTotalDistanceUpdatedAt()) {
                $ownerChair->setTotalDistanceUpdatedAt($row->totalDistanceUpdatedAtUnixMilliseconds());
            }
            $ownerChairs[] = $ownerChair;
        }
        return $this->writeJson($response, $res->setChairs($ownerChairs));
    }
}
