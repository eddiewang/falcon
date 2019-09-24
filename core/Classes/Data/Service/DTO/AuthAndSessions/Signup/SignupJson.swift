//
//  SignUp.swift
//  falcon
//
//  Created by Manu Herrera on 24/08/2018.
//  Copyright © 2018 muun. All rights reserved.
//

struct SignupJson: Codable {

    let firstName: String?
    let lastName: String?
    let email: String?
    let primaryCurrency: String
    let basePublicKey: PublicKeyJson
    let passwordChallengeSetup: ChallengeSetupJson

}
