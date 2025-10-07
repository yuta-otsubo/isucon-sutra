import { FC } from "react";
import { Loading } from "~/components/icon/loading";
import { LocationButton } from "~/components/modules/location-button/location-button";
import { PriceText } from "~/components/modules/price-text/price-text";
import { Text } from "~/components/primitives/text/text";
import { Coordinate } from "~/types";

export const Matching: FC<{
  pickup?: Coordinate;
  destLocation?: Coordinate;
  fare?: number;
}> = ({ pickup, destLocation, fare }) => {
  return (
    <div className="w-full h-full px-8 flex flex-col items-center justify-center">
      <Loading className="mb-8" />
      <Text size="xl" className="mb-6">
        マッチングしています
      </Text>
      <LocationButton
        label="現在地"
        location={pickup}
        className="w-80"
        disabled
      />
      <Text size="xl">↓</Text>
      <LocationButton
        label="目的地"
        location={destLocation}
        className="w-80"
        disabled
      />
      <p className="mt-8">
        {typeof fare === "number" ? (
          <>
            予定運賃: <PriceText tagName="span" value={fare} />
          </>
        ) : null}
      </p>
    </div>
  );
};
