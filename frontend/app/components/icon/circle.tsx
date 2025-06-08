import type { IconType } from "~/types";

/**
 * path: 円形のアウトラインを描画するための命令
 * 14z までで外形の円を、16z までで内側の円を描画
 */
export const CircleIcon: IconType = function (props) {
  return (
    <svg
      fill="currentColor"
      viewBox="0 0 16 16"
      height="1em"
      width="1em"
      {...props}
    >
      <path d="M8 15A7 7 0 118 1a7 7 0 010 14zm0 1A8 8 0 108 0a8 8 0 000 16z" />
    </svg>
  );
};
