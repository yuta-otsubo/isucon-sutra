// コンテンツを表示するUIコンポーネントで、既存のコンテンツの上に重ねて表示される
import { FC, PropsWithChildren }  from 'react';

type ModalProps = PropsWithChildren<{

}>;

export const ButtonLink: FC<ModalProps> = ({ children }) => {
    return (
        <div
            className=""
        >
            {children}
        </div>
    );
};
