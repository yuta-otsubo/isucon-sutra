import { useState } from "react";
import { CarRedIcon } from "~/components/icon/car-red";
import { LocationButton } from "~/components/modules/location-button/location-button";
import { Map } from "~/components/modules/map/map";
import { Button } from "~/components/primitives/button/button";
import { Text } from "~/components/primitives/text/text";
import type { RequestProps } from "~/components/request/type";
import { ReceptionMapModal } from "./receptionMapModal";

type Action = "from" | "to";

export const Reception = ({
  status,
}: RequestProps<"IDLE" | "MATCHING" | "DISPATCHING">) => {
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [action, setAction] = useState<Action>("from");

  const handleOpenModal = (action: Action) => {
    setIsModalOpen(true);
    setAction(action);
  };

  const onCloseModal = () => {
    setIsModalOpen(false);
  };

  return (
    <>
      {status === "IDLE" ? (
        <>
          <Map />
          <div className="w-full px-8 py-8 flex flex-col items-center justify-center">
            <LocationButton
              type="from"
              position="here"
              className="w-full"
              onClick={() => handleOpenModal("from")}
            />
            <Text size="xl">↓</Text>
            <LocationButton
              type="to"
              position={{ latitude: 123, longitude: 456 }}
              className="w-full"
              onClick={() => handleOpenModal("to")}
            />
            <Button
              variant="primary"
              className="w-full mt-6 font-bold"
              onClick={() => {}}
            >
              ISURIDE
            </Button>
          </div>
        </>
      ) : (
        <div className="w-full h-full px-8 flex flex-col items-center justify-center">
          <CarRedIcon className="size-[76px] mb-4" />
          <Text size="xl" className="mb-6">
            配車しています
          </Text>
          <LocationButton
            type="from"
            position="here"
            className="w-full"
            onClick={() => handleOpenModal("from")}
          />
          <Text size="xl">↓</Text>
          <LocationButton
            type="to"
            position={{ latitude: 123, longitude: 456 }}
            className="w-full"
            onClick={() => handleOpenModal("to")}
          />
          <Button variant="danger" className="w-full mt-6" onClick={() => {}}>
            配車をキャンセルする
          </Button>
        </div>
      )}

      {isModalOpen && (
        <ReceptionMapModal onClose={onCloseModal}>
          {action === "from" ? "この場所から移動する" : "この場所に移動する"}
        </ReceptionMapModal>
      )}
    </>
  );
};
