import type { ComponentProps, FC } from "react";

export const ChairIcon: FC<ComponentProps<"svg">> = function (props) {
  return (
    <svg
      fill="currentColor"
      viewBox="0 0 16 16"
      height="1em"
      width="1em"
      {...props}
    >
      {/* 仮 */}
      <path d="M2 2 L14 2 L14 14 L2 14 Z" />
    </svg>
  );
};
