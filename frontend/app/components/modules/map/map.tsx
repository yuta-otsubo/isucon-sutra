import {
  ComponentProps,
  FC,
  MouseEventHandler,
  TouchEventHandler,
  useCallback,
  useEffect,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import { twMerge } from "tailwind-merge";
import colors from "tailwindcss/colors";
import { ChairIcon } from "~/components/icon/chair";
import { PinIcon } from "~/components/icon/pin";
import { Button } from "~/components/primitives/button/button";
import type { Coordinate, DisplayPos, NearByChair } from "~/types";

const GridDistance = 20;
const PinSize = 50;
const ChairSize = 40;
const DisplayMapSize = GridDistance * 80;
const WorldSize = 1000;

const minmax = (num: number, min: number, max: number) => {
  return Math.min(Math.max(num, min), max);
};

const coordinateToPos = ({ latitude, longitude }: Coordinate): DisplayPos => {
  return {
    x: (-(latitude + WorldSize / 2) / WorldSize) * DisplayMapSize,
    y: (-(longitude + WorldSize / 2) / WorldSize) * DisplayMapSize,
  };
};

const posToCoordinate = ({ x, y }: DisplayPos): Coordinate => {
  return {
    latitude: Math.ceil((-x / DisplayMapSize) * WorldSize - WorldSize / 2),
    longitude: Math.ceil((-y / DisplayMapSize) * WorldSize - WorldSize / 2),
  };
};

const centerPosFrom = (pos: DisplayPos, outerRect: DOMRect): DisplayPos => {
  return {
    x: pos.x - outerRect.width / 2,
    y: pos.y - outerRect.height / 2,
  };
};

const draw = (
  ctx: CanvasRenderingContext2D,
  option: { from?: Coordinate; to?: Coordinate },
) => {
  // background
  ctx.fillStyle = colors.neutral[100];
  ctx.fillRect(0, 0, DisplayMapSize, DisplayMapSize);

  ctx.strokeStyle = colors.neutral[200];
  ctx.lineWidth = 2;
  ctx.lineCap = "round";
  ctx.setLineDash([]);
  ctx.beginPath();
  for (let v = GridDistance; v < DisplayMapSize; v += GridDistance) {
    ctx.moveTo(v, 0);
    ctx.lineTo(v, DisplayMapSize);
  }
  for (let h = GridDistance; h < DisplayMapSize; h += GridDistance) {
    ctx.moveTo(0, h);
    ctx.lineTo(DisplayMapSize, h);
  }
  ctx.stroke();

  // from-to
  const from = option.from ? coordinateToPos(option.from) : undefined;
  const to = option.to ? coordinateToPos(option.to) : undefined;

  if (from && to) {
    ctx.strokeStyle = colors.neutral[400];
    ctx.lineWidth = 3;
    ctx.lineCap = "round";
    ctx.setLineDash([3, 12]);
    ctx.beginPath();
    ctx.moveTo(-from.x, -from.y);
    ctx.lineTo(-to.x, -to.y);
    ctx.stroke();
  }

  if (from) {
    ctx.fillStyle = colors.neutral[800];
    ctx.beginPath();
    ctx.arc(-from.x, -from.y, 3, 0, 2 * Math.PI);
    ctx.fill();
  }

  if (to) {
    ctx.fillStyle = colors.red[500];
    ctx.beginPath();
    ctx.arc(-to.x, -to.y, 3, 0, 2 * Math.PI);
    ctx.fill();
  }
};

const SelectorLayer: FC<{
  pinSize?: number;
  pinColor?: `#${string}`;
  pos?: DisplayPos;
  updateViewLocation: (coordinate: Coordinate) => void;
}> = ({ pinSize = 80, pinColor = colors.black, pos, updateViewLocation }) => {
  const loc = useMemo(() => pos && posToCoordinate(pos), [pos]);
  const [isOpenCustomSelector, setIsOpenCustomSelector] = useState(false);
  const inputLatitudeRef = useRef<HTMLInputElement>(null);
  const inputLongitudeRef = useRef<HTMLInputElement>(null);

  return (
    <div className="flex items-center justify-center w-full h-full">
      <svg
        className="absolute top-0 left-0 w-full h-full opacity-10"
        xmlns="http://www.w3.org/2000/svg"
      >
        <rect x="50%" y="0" width={2} height={"100%"} />
        <rect x="0" y="50%" width={"100%"} height={2} />
      </svg>
      <PinIcon
        className="absolute mt-[-8px] opacity-60"
        color={pinColor}
        width={pinSize}
        height={pinSize}
        style={{
          transform: `translateY(-${pinSize / 2}px)`,
        }}
      />
      {loc && (
        <div className="absolute right-6 bottom-5 text-neutral-500 font-mono">
          <span>{`${loc.latitude}, ${loc.longitude}`}</span>
        </div>
      )}
      <Button
        className="py-2 px-3 absolute left-4 bottom-4"
        onClick={() => setIsOpenCustomSelector(true)}
      >
        Custom
      </Button>
      {isOpenCustomSelector && loc && (
        <div className="p-4 bg-neutral-50 bg-opacity-80 absolute top-0 left-0 w-full h-full flex items-center justify-center flex-col">
          <div className="flex space-x-4 w-full">
            <div className="mb-3 flex-1">
              <label htmlFor="latitude" className="block text-neutral-600 mb-1">
                Latitude:
              </label>
              <input
                type="number"
                id="latitude"
                min={0}
                max={DisplayMapSize}
                defaultValue={loc.latitude}
                placeholder="latitude"
                className="px-3 py-2 w-full border border-neutral-300 rounded focus:outline-none focus:ring-1 focus:ring-neutral-400"
                ref={inputLatitudeRef}
              />
            </div>
            <div className="mb-3 flex-1">
              <label
                htmlFor="longtiude"
                className="block text-neutral-600 mb-1"
              >
                Longitude:
              </label>
              <input
                type="number"
                id="longtiude"
                min={0}
                max={DisplayMapSize}
                defaultValue={loc.longitude}
                placeholder="longitude"
                className="px-3 py-2 w-full border border-neutral-300 rounded focus:outline-none focus:ring-1 focus:ring-neutral-400"
                ref={inputLongitudeRef}
              />
            </div>
          </div>
          <Button
            onClick={() => {
              const inputLoc = {
                latitude: Number(inputLatitudeRef.current?.value ?? 0),
                longitude: Number(inputLongitudeRef.current?.value ?? 0),
              };
              updateViewLocation(inputLoc);
              setIsOpenCustomSelector(false);
            }}
          >
            位置をセット
          </Button>
        </div>
      )}
    </div>
  );
};

const PinLayer: FC<{
  from?: Coordinate;
  to?: Coordinate;
}> = ({ from, to }) => {
  const fromPos = useMemo(() => from && coordinateToPos(from), [from]);
  const toPos = useMemo(() => to && coordinateToPos(to), [to]);
  return (
    <div className="flex w-full h-full absolute top-0 left-0">
      {fromPos && (
        <PinIcon
          className="absolute top-0 left-0 transition-transform duration-300 ease-in-out"
          color={colors.black}
          width={PinSize}
          height={PinSize}
          style={{
            transform: `translate(${-fromPos.x - PinSize / 2}px, ${-fromPos.y - PinSize - 8}px)`,
          }}
        />
      )}
      {toPos && (
        <PinIcon
          className="absolute top-0 left-0 transition-transform duration-300 ease-in-out"
          color={colors.red[500]}
          width={PinSize}
          height={PinSize}
          style={{
            transform: `translate(${-toPos.x - PinSize / 2}px, ${-toPos.y - PinSize - 8}px)`,
          }}
        ></PinIcon>
      )}
    </div>
  );
};

const ChairLayer: FC<{
  chairs?: NearByChair[];
}> = ({ chairs }) => {
  return (
    <div className="flex w-full h-full absolute top-0 left-0">
      {chairs?.map(({ id, model, current_coordinate }) => {
        const pos = coordinateToPos(current_coordinate);
        return (
          <ChairIcon
            model={model}
            key={id}
            width={ChairSize}
            height={ChairSize}
            className="absolute top-0 left-0 transition-transform duration-300 ease-in-out"
            style={{
              transform: `translate(${-pos.x - PinSize / 2}px, ${-pos.y - PinSize - 8}px)`,
            }}
          />
        );
      })}
    </div>
  );
};

type MapProps = ComponentProps<"div"> & {
  onMove?: (coordinate: Coordinate) => void;
  selectable?: boolean;
  selectorPinColor?: `#${string}`;
  from?: Coordinate;
  to?: Coordinate;
  chairs?: NearByChair[];
  initialCoordinate?: Coordinate;
};

export const Map: FC<MapProps> = ({
  selectable,
  selectorPinColor,
  onMove,
  from,
  to,
  chairs,
  initialCoordinate,
  className,
}) => {
  const onMoveRef = useRef(onMove);
  const outerRef = useRef<HTMLDivElement>(null);
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const [isDrag, setIsDrag] = useState(false);
  const [{ x, y }, setPos] = useState({ x: 0, y: 0 });
  const [movingStartPos, setMovingStartPos] = useState({ x: 0, y: 0 });
  const [movingStartPagePos, setMovingStartPagePos] = useState({
    x: 0,
    y: 0,
  });
  const [outerRect, setOuterRect] = useState<DOMRect | undefined>(undefined);

  const updateViewLocation = useCallback((loc?: Coordinate) => {
    if (!outerRef.current) {
      return;
    }
    const rect = outerRef.current.getBoundingClientRect();
    if (loc) {
      const pos = coordinateToPos(loc);
      const initalPos = {
        x: pos.x + rect.width / 2,
        y: pos.y + rect.height / 2,
      };
      setPos(initalPos);
      onMoveRef?.current?.(loc);
      return;
    }
    if (!loc) {
      const mapCenterPos = {
        x: -DisplayMapSize / 2 + rect.width / 2,
        y: -DisplayMapSize / 2 + rect.height / 2,
      };
      setPos(mapCenterPos);
      onMoveRef?.current?.(posToCoordinate(centerPosFrom(mapCenterPos, rect)));
      return;
    }
  }, []);

  useLayoutEffect(() => {
    updateViewLocation(initialCoordinate);
  }, [initialCoordinate, updateViewLocation]);

  useEffect(() => {
    const canvas = canvasRef.current;
    const context = canvas?.getContext("2d");
    if (!context) return;
    draw(context, { from, to });
  }, [from, to]);

  useEffect(() => {
    const observer = new ResizeObserver((entries) => {
      setOuterRect(entries[0].contentRect);
    });
    if (outerRef.current) {
      observer.observe(outerRef.current);
    }
    return () => {
      observer.disconnect();
    };
  }, []);

  const onMouseDown: MouseEventHandler<HTMLDivElement> = useCallback(
    (e) => {
      setIsDrag(true);
      setMovingStartPagePos({ x: e.pageX, y: e.pageY });
      setMovingStartPos({ x, y });
    },
    [x, y],
  );

  const onTouchStart: TouchEventHandler<HTMLDivElement> = useCallback(
    (e) => {
      setIsDrag(true);
      setMovingStartPagePos({
        x: e.touches[0].pageX,
        y: e.touches[0].pageY,
      });
      setMovingStartPos({ x, y });
    },
    [x, y],
  );

  useEffect(() => {
    const onEnd = () => {
      setMovingStartPagePos({ x: 0, y: 0 });
      setIsDrag(false);
    };
    window.addEventListener("mouseup", onEnd);
    window.addEventListener("touchend", onEnd);
    return () => {
      window.removeEventListener("mouseup", onEnd);
      window.removeEventListener("touchend", onEnd);
    };
  }, []);

  useEffect(() => {
    const setFixedPos = (pageX: number, pageY: number) => {
      if (!outerRect) return;
      const posX = minmax(
        movingStartPos.x - (movingStartPagePos.x - pageX),
        -DisplayMapSize + outerRect.width,
        0,
      );
      const posY = minmax(
        movingStartPos.y - (movingStartPagePos.y - pageY),
        -DisplayMapSize + outerRect.height,
        0,
      );
      setPos({ x: posX, y: posY });
      onMoveRef?.current?.(
        posToCoordinate(centerPosFrom({ x: posX, y: posY }, outerRect)),
      );
    };
    const onMouseMove = (e: MouseEvent) => {
      setFixedPos(e.pageX, e.pageY);
    };
    const onTouchMove = (e: TouchEvent) => {
      setFixedPos(e.touches[0].pageX, e.touches[0].pageY);
    };
    if (isDrag) {
      window.addEventListener("mousemove", onMouseMove, { passive: true });
      window.addEventListener("touchmove", onTouchMove, { passive: true });
    }
    return () => {
      window.removeEventListener("mousemove", onMouseMove);
      window.removeEventListener("touchmove", onTouchMove);
    };
  }, [isDrag, movingStartPagePos, movingStartPos, onMove, outerRect]);

  return (
    <div
      className={twMerge(
        "w-full h-full relative overflow-hidden bg-neutral-200",
        isDrag && "cursor-grab",
        className,
      )}
      ref={outerRef}
      onMouseDown={onMouseDown}
      onTouchStart={onTouchStart}
      role="button"
      tabIndex={0}
    >
      <div
        className={twMerge(
          "absolute top-0 left-0",
          !isDrag && "transition-transform duration-200 ease-in-out",
        )}
        style={{
          transform: `translate(${x}px, ${y}px)`,
          width: DisplayMapSize,
          height: DisplayMapSize,
        }}
      >
        <canvas
          width={DisplayMapSize}
          height={DisplayMapSize}
          ref={canvasRef}
        />
        {chairs && chairs.length !== 0 && <ChairLayer chairs={chairs} />}
        <PinLayer from={from} to={to} />
      </div>
      {selectable && outerRect && (
        <SelectorLayer
          pos={centerPosFrom({ x, y }, outerRect)}
          updateViewLocation={updateViewLocation}
          pinColor={selectorPinColor}
        />
      )}
    </div>
  );
};
