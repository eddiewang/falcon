//
//  AddressActions.swift
//  falcon
//
//  Created by Juan Pablo Civile on 10/12/2018.
//  Copyright © 2018 muun. All rights reserved.
//

import Foundation
import RxSwift
import Libwallet

public class AddressActions {

    private let houstonService: HoustonService
    private let keysRepository: KeysRepository
    private let syncExternalAddresses: SyncExternalAddresses

    init(houstonService: HoustonService,
         keysRepository: KeysRepository,
         syncExternalAddresses: SyncExternalAddresses) {

        self.houstonService = houstonService
        self.keysRepository = keysRepository
        self.syncExternalAddresses = syncExternalAddresses
    }

    func syncPublicKeySet() -> Completable {

        return Single.deferred { () throws -> Single<PublicKeySet> in
                let publicKey = try self.keysRepository.getBasePublicKey()
                return self.houstonService.updatePublicKeySet(publicKey: publicKey)
            }
            .do(onSuccess: { (keySet) in

                if let indices = keySet.externalPublicKeyIndices {
                    self.keysRepository.updateMaxUsedIndex(indices.maxUsedIndex)

                    if let maxWatchIndex = indices.maxWatchingIndex {
                        self.keysRepository.updateMaxWatchingIndex(maxWatchIndex)
                    }
                }

                if let muunPublicKey = keySet.baseCosigningPublicKey {
                    self.keysRepository.store(cosigningKey: muunPublicKey)
                }

            })
            .asCompletable()
    }

    public func generateExternalAddress() throws -> LibwalletMuunAddressProtocol {

        let maxUsedIndex = keysRepository.getMaxUsedIndex()
        let maxWatchingIndex = keysRepository.getMaxWatchingIndex()

        let nextIndex: Int
        if maxUsedIndex < 0 {
            nextIndex = 0
        } else if maxUsedIndex < maxWatchingIndex {
            nextIndex = maxUsedIndex + 1
        } else {
            nextIndex = Int.random(in: 0 ..< maxWatchingIndex)
        }

        let derivedMuunKey = try keysRepository.getCosigningKey()
            .derive(to: .external)
            .derive(at: UInt32(nextIndex))

        let derivedKey = try keysRepository.getBasePublicKey()
            .derive(to: .external)
            .derive(at: UInt32(nextIndex))

        let address = try doWithError { error in
            LibwalletCreateAddressV3(derivedKey.key, derivedMuunKey.key, error)
        }

        if nextIndex > maxUsedIndex {
            keysRepository.updateMaxUsedIndex(nextIndex)

            syncExternalAddresses.run()
        }

        return address
    }
}
